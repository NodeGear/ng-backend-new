package nodegear

import (
	"fmt"
	"time"
	"../models"
	"../connection"
	"gopkg.in/mgo.v2/bson"
)

func (p *Instance) DidStart() {
	fmt.Println("Did Start")
	p.Running = true
	p.Starting = false

	p.Intended_stop = false

	(&models.AppEvent{
		App: p.App_id,
		Process: p.Process_id,
		Name: "Start",
		Message: "Process has been started",
	}).Add()

	p.AddToProxy()

	uptime := p.GetUptime()
	uptime.Start = time.Now()
	c := connection.MongoC(models.AppProcessUptimeC)
	err := c.UpdateId(uptime.ID, &bson.M{
		"$set": &bson.M{
			"start": uptime.Start,
		},
	})

	if err != nil {
		panic(err)
	}

	c = connection.MongoC(models.AppProcessC)
	err = c.UpdateId(p.Process_id, &bson.M{
		"$set": &bson.M{
			"running": true,
			"containerID": p.Container_id,
		},
	})

	if err != nil {
		panic(err)
	}

	redis := connection.Redis().Get()
	if _, err := redis.Do("PUBLISH", "pm:app_running", p.Process_id.Hex() + "|true"); err != nil {
		panic(err)
	}
	redis.Close()
}

func (p *Instance) DidStop() {
	fmt.Println("Did Stop")
	p.Running = false
	p.Starting = false

	(&models.AppEvent{
		App: p.App_id,
		Process: p.Process_id,
		Name: "Stop",
		Message: "Process stopped",
	}).Add()

	p.RemoveFromProxy()

	c := connection.MongoC(models.AppProcessC)
	err := c.UpdateId(p.Process_id, &bson.M{
		"$set": &bson.M{
			"running": false,
			"containerID": nil,
		},
	})

	if err != nil {
		panic(err)
	}

	redis := connection.Redis().Get()
	if _, err := redis.Do("PUBLISH", "pm:app_running", p.Process_id.Hex() + "|false"); err != nil {
		panic(err)
	}
	redis.Close()
}

func (p *Instance) DidExit() {
	fmt.Println("Did Exit")
	p.Running = false
	p.Starting = false

	(&models.AppEvent{
		App: p.App_id,
		Process: p.Process_id,
		Name: "Exit",
		Message: "Process quit",
	}).Add()

	p.RemoveFromProxy()

	c := connection.MongoC(models.AppProcessC)
	err := c.UpdateId(p.Process_id, &bson.M{
		"$set": &bson.M{
			"running": false,
			"containerID": nil,
		},
	})

	if err != nil {
		panic(err)
	}

	redis := connection.Redis().Get()
	if _, err := redis.Do("PUBLISH", "pm:app_running", p.Process_id.Hex() + "|false"); err != nil {
		panic(err)
	}
	redis.Close()

	p.Remove()
}

func (p *Instance) DidRestart() {
	fmt.Println("Did Restart")
}
