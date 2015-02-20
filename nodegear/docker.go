package nodegear

import (
	//"../connection"
	"../docker"
	"../config"
	"../models"
	"net/http"
	"fmt"
	"encoding/json"
	"bytes"
	"bufio"
	"io"
	"io/ioutil"
	"strconv"
)

type createContainer struct {
	Hostname string `json:"Hostname"`
	Domainname string `json:"Domainname"`
	User string `json:"User"`
	Memory int `json:"Memory"`
	MemorySwap int `json:"MemorySwap"`
	CpuShares int `json:"CpuShares"`
	Cpuset string `json:"Cpuset"`
	AttachStdin bool `json:"AttachStdin"`
	AttachStdout bool `json:"AttachStdout"`
	AttachStderr bool `json:"AttachStderr"`
	PortSpecs map[string]interface{} `json:"PortSpecs"`
	ExposedPorts map[string]interface{} `json:"ExposedPorts"`
	Tty bool `json:"Tty"`
	OpenStdin bool `json:"OpenStdin"`
	StdinOnce bool `json:"StdinOnce"`
	Env []string `json:"Env"`
	Cmd *string `json:"Cmd"`
	Image string `json:"Image"`
	Volumes struct{} `json:"Volumes"`
	WorkingDir string `json:"WorkingDir"`
	Entrypoint *string `json:"Entrypoint"`
	NetworkDisabled bool `json:"NetworkDisabled"`
	OnBuild map[string]interface{} `json:"OnBuild"`
}

type portBinding struct {
	HostIp string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

type startContainer struct {
	Binds []string `json:"Binds"`
	ContainerIDFile string `json:"ContainerIDFile"`
	LxcConf []string `json:"LxcConf"`
	Privileged bool `json:"Privileged"`
	PortBindings map[string][]*portBinding `json:"PortBindings"`
	Links []string `json:"Links"`
	PublishAllPorts bool `json:"PublishAllPorts"`
	Dns []string `json:"Dns"`
	DnsSearch *string `json:"DnsSearch"`
	VolumesFrom *string `json:"VolumesFrom"`
	NetworkMode string `json:"NetworkMode"`
	CapAdd *[]string `json:"CapAdd"`
	CapDrop *[]string `json:"CapDrop"`
	RestartPolicy struct {
		Name string `json:"Name"`
		MaximumRetryCount int `json:"MaximumRetryCount"`
	} `json:"RestartPolicy"`
}

type createContainerResponse struct {
	Id string `json:"Id"`
	Warnings []string `json:"Warnings"`
}

type DockerEvent struct {
	Status string `json:"status"`
	Id string `json:"id"`
	From string `json:"from"`
	Time int `json:"time"` // unix timestamp
}

type containerResponse struct {
	Id string `json:"Id"`
	State struct {
		Running bool `json:"Running"`
		Pid int `json:"Pid"`
		ExitCode int `json:"ExitCode"`
	} `json:"State"`
}

func ListenForEvents () {
	client := docker.GetClient()
	req, err := client.Get(config.Configuration.Docker_url + "/events")

	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(req.Body)
	for {
		var event DockerEvent

		if err := decoder.Decode(&event); err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		proc := FindInstanceByContainer(event.Id)

		if proc == nil {
			fmt.Println("Container", event.Id, "Not managed by nodegear!")
			continue
		}

		switch event.Status {
		case "start":
			proc.DidStart()
		case "stop":
			proc.DidStop()
		case "die":
			proc.DidExit()
		case "restart":
			proc.DidRestart()
		}

		fmt.Printf("Docker Event: %v, proc %v\n", event, proc)
	}
}

func (p *Instance) CreateContainer(environment []string, app *models.App) {
	createReq := &createContainer{
		Env: environment,
		ExposedPorts: map[string]interface{}{ "80/tcp": struct{}{}},
		Volumes: struct{}{},
		Image: app.Docker.Image,
		Cmd: &app.Docker.Command,
	}

	createReqBody, err := json.Marshal(createReq)
	if err != nil {
		panic(err)
	}

	client := docker.GetClient()
	request, err := http.NewRequest("POST", config.Configuration.Docker_url + "/v1.14/containers/create", bytes.NewBuffer(createReqBody))
	if err != nil {
		panic(err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Docker-Client/1.2.0")

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	fmt.Println("Create container response", response.Status)
	if response.StatusCode == 404 {
		p.Log("\n [WARN] Container not found. Pulling container.\n")
		didPull := p.PullContainer(app)

		if !didPull {
			p.Log("\n [ERR] Could not pull container.\n")
			p.Starting = false
			p.Running = false

			p.DidExit()
			return
		}

		request, err = http.NewRequest("POST", config.Configuration.Docker_url + "/v1.14/containers/create", bytes.NewBuffer(createReqBody))
		if err != nil {
			panic(err)
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("User-Agent", "Docker-Client/1.2.0")

		response, err = client.Do(request)
		if err != nil {
			panic(err)
		}
	}

	if response.StatusCode != 201 {
		p.Log("\n [ERR] Could not create container. This is an internal error, please contact us or try again to resolve this.\n")
		p.Starting = false
		p.Running = false

		p.DidExit()
		return
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var createResponse createContainerResponse
	if err = json.Unmarshal(body, &createResponse); err != nil {
		panic(err)
	}

	p.Container_id = createResponse.Id
	fmt.Println("Container:", p.Container_id)
}

func (p *Instance) StartContainer() {
	userpath := config.Configuration.Homepath + p.User_id.Hex()

	var binding []*portBinding
	binding = append(binding, &portBinding{
		HostIp: "",
		HostPort: strconv.Itoa(p.Port),
	})

	bind := make(map[string][]*portBinding)
	bind["80/tcp"] = binding

	start := startContainer{
		Binds: []string{userpath + "/" + p.Process_id.Hex() + ":/srv/app:rw", userpath + "/.ssh:/root/.ssh:r"},
		PortBindings: bind,
		Dns: []string{"8.8.8.8", "8.8.4.4"},
	}

	reqBody, err := json.Marshal(start)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(reqBody))

	client := docker.GetClient()
	request, err := http.NewRequest("POST", config.Configuration.Docker_url + "/v1.14/containers/" + p.Container_id + "/start", bytes.NewBuffer(reqBody))
	if err != nil {
		panic(err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Docker-Client/1.2.0")

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	fmt.Println(response.Status)
	if response.StatusCode != 204 {
		p.Log("\n [ERR] Could not start container. This is an internal error, please contact us or try again to resolve this.\n")
		p.Starting = false
		p.Running = false

		p.DidExit()
		return
	}

	p.Log("\n\n Application Started!\n ====================\n")

	go p.GetContainerLogs()
}

func (p *Instance) StopContainer() {
	client := docker.GetClient()
	request, err := http.NewRequest("POST", config.Configuration.Docker_url + "/v1.14/containers/" + p.Container_id + "/stop?t=5", bytes.NewBuffer([]byte("")))
	if err != nil {
		panic(err)
	}

	request.Header.Set("User-Agent", "Docker-Client/1.2.0")

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	fmt.Println(response.Status)
	if response.StatusCode != 204 {
		p.Log("\n [ERR] Could not stop container.\n")
		p.Starting = false
		p.Running = false

		p.DidExit()
		return
	}

	p.Log("\n Application Stopped.\n ====================\n\n")
}

func (p *Instance) GetContainerLogs() {
	client := docker.GetClient()
	request, err := http.NewRequest("GET", config.Configuration.Docker_url + "/v1.14/containers/" + p.Container_id + "/logs?stderr=1&stdout=1&follow=1", bytes.NewBuffer([]byte("")))
	if err != nil {
		panic(err)
	}

	request.Header.Set("User-Agent", "Docker-Client/1.2.0")

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	fmt.Println(response.Status)

	reader := bufio.NewScanner(response.Body)
	for reader.Scan() {
		line := reader.Bytes()
		if len(line) > 8 {
			line = line[8:]
		}

		p.Log(string(line) + "\n")
	}

	if err := reader.Err(); err != nil {
		panic(err)
	}

	fmt.Println("Quit trailing loggs")
}

func (p *Instance) PullContainer(app *models.App) bool {
	client := docker.GetClient()
	request, err := http.NewRequest("POST", config.Configuration.Docker_url + "/v1.14/images/create?fromImage=" + app.Docker.Image, bytes.NewBuffer([]byte("")))
	if err != nil {
		panic(err)
	}

	request.Header.Set("User-Agent", "Docker-Client/1.2.0")

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	fmt.Println(response.Status)
	if response.StatusCode != 200 {
		return false
	}

	reader := bufio.NewReader(response.Body)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else {
			panic(err)
		}

		fmt.Println("Pull Image Response", line)
	}

	return true
}

func (p *Instance) GetContainer() *containerResponse {
	client := docker.GetClient()
	request, err := http.NewRequest("GET", config.Configuration.Docker_url + "/v1.14/containers/" + p.Container_id + "/json", bytes.NewBuffer([]byte("")))
	if err != nil {
		panic(err)
	}

	request.Header.Set("User-Agent", "Docker-Client/1.2.0")

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	fmt.Println("Get docker info response status", response.Status)
	if response.StatusCode != 200 {
		return nil
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var res containerResponse
	if err = json.Unmarshal(body, &res); err != nil {
		panic(err)
	}

	return &res
}

func (p *Instance) DeleteContainer() {
	if len(p.Container_id) == 0 {
		return
	}

	client := docker.GetClient()
	request, err := http.NewRequest("DELETE", config.Configuration.Docker_url + "/v1.14/containers/" + p.Container_id + "?v=0", bytes.NewBuffer([]byte("")))
	if err != nil {
		panic(err)
	}

	request.Header.Set("User-Agent", "Docker-Client/1.2.0")

	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	fmt.Println("Delete container status:", response.Status)
	if response.StatusCode != 204 {
		p.Log("\n [ERR] Could not delete container.\n")

		return
	}

	p.Log("\n Container deleted.\n ====================\n\n")
}
