package routes

import (
	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"encoding/json"
	"../connection"
	"../nodegear"
	"../config"
	"../models"
)

type redisMsg struct {
	Id string `json:"id"`
	Action string `json:"action"`
}

func ListenToRedis () {
	pubsub_session := connection.Redis().Get()

	pub_sub := redis.PubSubConn{pubsub_session}
	pub_sub.Subscribe("server_" + config.Configuration.Server.Identifier)

	for {
		switch v := pub_sub.Receive().(type) {
			case redis.Message:
				go dispatch(v.Data)
			case error:
				fmt.Printf("Redis PubSub Error: %v", v)
		}
	}
}

func dispatch(msg []byte) {
	var message redisMsg
	if err := json.Unmarshal(msg, &message); err != nil {
		panic(err)
	}

	process := nodegear.FindInstanceByProcessId(bson.ObjectIdHex(message.Id))

	if process == nil {
		fmt.Println("Instance not found", message)
		
		c := connection.MongoC(models.AppProcessC)

		// Find the instance
		var db_process models.AppProcess
		err := c.FindId(bson.ObjectIdHex(message.Id)).One(&db_process)
		
		if err != nil {
			if err.Error() == "not found" {
				return
			} else {
				panic(err)
			}
		}

		process = &nodegear.Instance{
			Process_id: db_process.ID,
			App_id: db_process.App,
		}

		is := append(nodegear.Instances, process)
		nodegear.Instances = is
	}

	switch message.Action {
	case "start":
		go process.Start()
	case "stop":
		go process.Stop()
	case "restart_uptime":
		go process.RestartUptime()
	//case "restart":
	//	process.Restart()
	}
}
