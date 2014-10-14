package models

import "gopkg.in/mgo.v2/bson"
import "time"

var AppProcessC string = "appprocesses"

type AppProcess struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	
	Created time.Time `bson:"created"`
	Name string `bson:"name"`
	App bson.ObjectId `bson:"app,omitempty"`
	DataSnapshot bson.ObjectId `bson:"dataSnapshot,omitempty"`
	Server bson.ObjectId `bson:"server,omitempty"`
	Running bool `bson:"running"`
	Restarts bool `bson:"restarts"`
	Deleted bool `bson:"deleted"`
	ContainerID string `bson:"containerID"`
}
