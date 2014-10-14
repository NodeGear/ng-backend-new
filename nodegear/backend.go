package nodegear

import (
	"gopkg.in/mgo.v2/bson"
	"../models"
	"../connection"
	"../config"
	"fmt"
	"time"
)

var Server models.Server

func Init () {
	s := models.Server{
		Identifier: config.Configuration.Server.Identifier,
	}.FindByIdentifier()
	
	if s == nil {
		// I GOD, Me Creates Itself ... But someone must have created this god, right?
		fmt.Println("Creating new server")

		serverCol := connection.MongoC(models.ServerC)
		s = &models.Server{
			ID: bson.NewObjectId(),
			Created: time.Now(),
			Name: config.Configuration.Server.Name,
			Location: config.Configuration.Server.Location,
			Identifier: config.Configuration.Server.Identifier,
			Price_per_hour: config.Configuration.Server.Price_per_hour,
			App_memory: config.Configuration.Server.App_memory,
			AppLimit: config.Configuration.Server.AppLimit,
			AppsRunning: config.Configuration.Server.AppsRunning,
		}

		if err := serverCol.Insert(s); err != nil {
			panic(err)
		}

		redis := connection.Redis().Get()
		defer redis.Close()

		if _, err := redis.Do("SET", "backend_port:" + s.Identifier, 9000); err != nil {
			panic(err)
		}
	}

	Server = *s
}
