package nodegear

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/garyburd/redigo/redis"
	"../connection"
	"../models"
	"../config"
	"time"
	"fmt"
	"strconv"
	"os"
	"io/ioutil"
	"io"
	"os/exec"
	"bufio"
)

var Instances *[]*Instance

type Instance struct {
	App_id bson.ObjectId
	Process_id bson.ObjectId
	User_id bson.ObjectId

	RestartProcess bool
	CurrentLog string
	Inserted_log_to_redis bool

	App_location string
	Intended_stop bool

	Running bool
	Starting bool

	Uptime bson.ObjectId
	Start_time time.Time

	Port int

	Container_id string

	LastCPU uint64
	AppMemory uint64
}

func init() {
	is := make([]*Instance, 0)
	Instances = &is
}

func (p *Instance) Remove() {
	p.CleanProcess()

	for i, proc := range *Instances {
		if proc.Process_id.Hex() == p.Process_id.Hex() {
			instances := *Instances
			instances[i] = instances[len(instances)-1]
			instances = instances[:len(instances)-1]
			Instances = &instances
			return
		}
	}

	fmt.Println("not found")
}

func (p *Instance) GetAppModel(select_data *bson.M) *models.App {
	c := connection.MongoC(models.AppC)

	var app models.App

	q := c.FindId(p.App_id)

	if select_data != nil {
		q = q.Select(select_data)
	}

	err := q.One(&app)
	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		return nil
	} else if err != nil {
		panic(err)
	}

	return &app
}

func (p *Instance) GetAppProcessModel() *models.AppProcess {
	c := connection.MongoC(models.AppProcessC)

	var appProcess models.AppProcess
	err := c.FindId(p.Process_id).One(&appProcess)

	if err != nil && err.Error() == mgo.ErrNotFound.Error() {
		return nil
	} else if err != nil {
		panic(err)
	}

	return &appProcess
}

func (p *Instance) Init() bool {
	// Get the app
	app := p.GetAppModel(&bson.M{
		"user": 1,
	})

	if app == nil {
		return false
	}

	p.User_id = app.User
	p.App_location = config.Configuration.Homepath + p.User_id.Hex() + "/" + p.Process_id.Hex()
	p.AppMemory = uint64(config.Configuration.Server.App_memory) * 1024 * 1024

	return true
}

func (p *Instance) Launch() {
	p.Start_time = time.Now()
	p.Inserted_log_to_redis = false

	// Get the environment
	var env []models.AppEnvironment
	envc := connection.MongoC(models.AppEnvironmentC)

	envc.Find(&bson.M{
		"app": p.App_id,
	}).All(&env)

	redis_conn := connection.Redis().Get()

	port, err := redis.Int(redis_conn.Do("INCR", "backend_port:" + config.Configuration.Server.Identifier))
	if err != nil {
		panic(err)
	}

	redis_conn.Close()

	fmt.Println("Port:", port)
	p.Port = port

	environment := []string{"PORT=80", "NODE_ENV=production"}

	for _, e := range env {
		if e.Name == "DATABASE_HOST" && e.Value == "10.240.7.66" {
			e.Value = "10.240.249.130"
		}

		environment = append(environment, e.Name + "=" + e.Value)
	}

	fmt.Println(environment)

	p.CurrentLog = p.Process_id.Hex() + "_" + strconv.FormatInt(p.Start_time.UnixNano(), 10)

	// Get the app
	app := p.GetAppModel(&bson.M{
		"branch": 1,
		"script": 1,
		"user": 1,
		"location": 1,
		"app_type": 1,
		"docker": 1,
	})
	p.User_id = app.User

	// Install
	p.InstallProcess(app, environment)
}

func (p *Instance) Start() {
	if p.Running == true {
		(&models.AppEvent{
			App: p.App_id,
			Process: p.Process_id,
			Name: "Already Running",
			Message: "App is Already Running",
		}).Add()

		return
	}

	fmt.Println("Starting Process", p.Process_id)

	p.RestartProcess = false
	p.Intended_stop = false

	(&models.AppEvent{
		App: p.App_id,
		Process: p.Process_id,
		Name: "Starting",
		Message: "App is starting",
	}).Add()

	p.Launch()
}

func (p *Instance) Stop() {
	fmt.Println("Stopping Process", p.Process_id)

	p.RestartProcess = false
	p.Intended_stop = false

	(&models.AppEvent{
		App: p.App_id,
		Process: p.Process_id,
		Name: "Stopping",
		Message: "App is stopping",
	}).Add()

	if len(p.Container_id) > 0 {
		p.StopContainer()
		return
	}

	// Update process, set running to false
	c := connection.MongoC(models.AppProcessC)

	var processUpdate = bson.M{
		"$set": &bson.M{
			"running": false,
		},
	}
	if err := c.UpdateId(p.Process_id, &processUpdate); err != nil {
		panic(err)
	}

	redis_c := connection.Redis().Get()
	if _, err := redis_c.Do("PUBLISH", "pm:app_running", p.Process_id.Hex() + "|false"); err != nil {
		panic(err)
	}
	redis_c.Close()

	p.Remove()
}

func (p *Instance) InstallProcess(app *models.App, environment []string) {
	userpath := config.Configuration.Homepath + p.User_id.Hex()

	// Install the process
	if err := os.MkdirAll(userpath, 0755); err != nil {
		panic(err)
	}

	// Install private key under $HOME/.ssh
	p.InstallPrivateKey()

	// Download snapshot
	process := p.GetAppProcessModel()
	if process == nil {
		panic("Process NOT FOUND!")
	}

	if len(process.DataSnapshot.Hex()) > 0 {
		p.ApplySnapshot(process)
	}

	p.Log("\n Installation of App " + app.Name + " - Process " + process.Name + "\n ==========================================\n\n")

	if len(app.Location) == 0 {
		p.Log(" [ERR] Cannot Install: App does not have a valid git URL\n")
		p.Log(" [ERR] Invalid Git URL: " + app.Location + "\n")
		p.Stop()

		return
	}

	use_snapshot := "0"
	if len(process.DataSnapshot.Hex()) > 0 {
		use_snapshot = "1"
	}

	snapshot_path := "/tmp/snapshot_" + process.DataSnapshot.Hex() + ".diff"

	git_branch := app.Branch
	if len(git_branch) == 0 {
		git_branch = "master"
	}

	// Clones git, checks out the right branch and applies the snapshot
	command := exec.Command(config.Configuration.Scriptspath + "/installProcess.sh", userpath, p.App_location, app.Location, git_branch, use_snapshot, snapshot_path)
	reader, writer := io.Pipe()
	command.Stdout = writer
	command.Stderr = writer

	if err := command.Start(); err != nil {
		panic(err)
	}

	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			p.Log(" [Install] " + scanner.Text() + "\n")
		}
	}()

	if err := command.Wait(); err != nil {
		(&models.AppEvent{
			App: p.App_id,
			Process: p.Process_id,
			Name: "Install Error",
			Message: "App couldn't be installed at this time. Check the logs!",
		}).Add()

		p.Starting = false
		p.Running = false

		p.Stop()

		redis_conn := connection.Redis().Get()
		if _, err = redis_conn.Do("PUBLISH", "pm:app_running", p.Process_id.Hex() + "|false"); err != nil {
			panic(err)
		}
		redis_conn.Close()

		return
	}

	p.CreateContainer(environment, app)

	if len(p.Container_id) > 0 {
		p.StartContainer()
	} else {
		fmt.Println("Container ID not present. Somethign has gone wrong!", p)
	}
}

func (p *Instance) InstallPrivateKey() {
	userpath := config.Configuration.Homepath + p.User_id.Hex()

	var key models.RSAKey
	c := connection.MongoC(models.RSAKeyC)
	findErr := c.Find(&bson.M{
		"user": p.User_id.Hex(),
		"deleted": false,
		"installing": false,
		"installed": false,
		"system_key": true,
	}).Select(&bson.M{
		"private_key": 1,
		"public_key": 1,
	}).One(&key)

	if findErr != nil {
		if findErr.Error() == mgo.ErrNotFound.Error() {
			return
		}

		panic(findErr)
	}

	if err := os.MkdirAll(userpath + "/.ssh", 0700); err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(userpath + "/.ssh/id_rsa", []byte(key.Private_key), 0600); err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(userpath + "/.ssh/id_rsa.pub", []byte(key.Private_key), 0644); err != nil {
		panic(err)
	}

	ssh_config := "Host *\n" +
		"  StrictHostKeyChecking no\n" +
		"  CheckHostIp no\n" +
		"  PasswordAuthentication no\n"

	if err := ioutil.WriteFile(userpath + "/.ssh/config", []byte(ssh_config), 0644); err != nil {
		panic(err)
	}
}

func (p *Instance) CleanProcess() {
	fmt.Println("Cleaning Process", p.Process_id)

	p.DeleteContainer()

	use_snapshot := "0"
	snapshot_path := ""
	snapshot_id := bson.NewObjectId()

	app := p.GetAppModel(&bson.M{
		"useSnapshots": 1,
	})

	if config.Configuration.Storage.Enabled == true && app.UseSnapshots == true {
		use_snapshot = "1"
		snapshot_path = "/tmp/snapshot_" + snapshot_id.Hex() + ".diff"
	}

	fmt.Println(p.App_location)
	command := exec.Command(config.Configuration.Scriptspath + "/cleanProcess.sh", p.User_id.Hex(), p.App_location, use_snapshot, snapshot_path)

	if err := command.Start(); err != nil {
		panic(err)
	}

	if err := command.Wait(); err != nil {
		fmt.Println("Could not clean process", err)
		return
	}

	if config.Configuration.Storage.Enabled == true && app.UseSnapshots == true {
		p.SaveSnapshot(snapshot_id, snapshot_path)
	}
}
