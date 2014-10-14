package nodegear

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/garyburd/redigo/redis"
	"../models"
	"../connection"
	"fmt"
)

func FetchPreviousInstances () {
	c := connection.MongoC(models.AppProcessC)

	var procs []models.AppProcess

	c.Find(bson.M{
		"running": true,
		"server": Server.ID,
		"containerID": bson.M{
			"$ne": nil,
		},
	}).All(&procs)

	fmt.Println("Processes", procs)

	// for each, resurrect (ng-backend:ProcessManager.js:142)
	for _, proc := range procs {
		instance := FindInstanceByProcessId(proc.ID)

		if instance == nil {
			instance = &Instance{
				App_id: proc.App,
				Process_id: proc.ID,
			}

			Instances = append(Instances, instance)
		}

		userId := instance.GetAppModel(&bson.M{ "user": 1 }).User

		instance.Running = true
		instance.Inserted_log_to_redis = true
		instance.App_location = "/home/" + userId.Hex() + "/" + instance.Process_id.Hex()
		instance.Container_id = proc.ContainerID
		
		c := connection.MongoC(models.AppProcessUptimeC)
		var uptime models.AppProcessUptime

		q := &bson.M{
			"sealed": false,
			"user": userId,
			"app": instance.App_id,
			"process": instance.Process_id,
			"server": Server.ID,
		}

		if err := c.Find(q).One(&uptime); err != nil {
			panic(err)
		}

		redis_c := connection.Redis().Get()
		latest_log, err := redis.String(redis_c.Do("LINDEX", "pm:app_process_logs_" + instance.Process_id.Hex(), 0))
		redis_c.Close()
		
		if err != nil {
			panic(err)
		}

		if &latest_log == nil || &uptime == nil {
			// Never started
			instance.Remove()
			continue
		}

		instance.Uptime = uptime.ID
		instance.CurrentLog = latest_log

		container := instance.GetContainer()
		if container.State.Running != true {
			instance.Remove()
			continue
		}

		go instance.GetContainerLogs()
	}
}

func FindInstanceByProcessId (process_id bson.ObjectId) *Instance {
	for _, proc := range Instances {
		if proc.Process_id == process_id {
			return proc
		}
	}

	return nil
}

func FindInstanceByContainer (container_id string) *Instance {
	fmt.Println("Instances:", len(Instances))
	for _, proc := range Instances {
		fmt.Println("proc.Container_id", proc.Container_id)
		fmt.Println(container_id)
		if proc.Container_id == container_id {
			return proc
		}
	}

	return nil
}
