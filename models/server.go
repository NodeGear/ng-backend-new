package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
	"../connection"
)

var ServerC string = "servers"

type Server struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`

	Created time.Time `json:"created" bson:"created"`
	
	Name string `json:"name" bson:"name"`
	Location string `json:"location" bson:"location"`
	Identifier string `json:"identifier" bson:"identifier"`
	Price_per_hour float32 `json:"price_per_hour" bson:"price_per_hour"`
	
	App_memory int `json:"app_memory" bson:"app_memory"`
	AppLimit int `json:"appLimit" bson:"appLimit"`
	AppsRunning int `json:"appsRunning" bson:"appsRunning"`
}

func (s Server) FindByIdentifier () *Server {
	c := connection.MongoC(ServerC)

	var server Server
	q := bson.M{"identifier": s.Identifier}

	if err := c.Find(q).One(&server); err != nil {
		return nil
	}

	return &server
}