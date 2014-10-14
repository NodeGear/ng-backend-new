package models

import "gopkg.in/mgo.v2/bson"
import "time"

var AppEnvironmentC string = "appenvironments"

type AppEnvironment struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	
	Created time.Time `bson:"created"`
	App bson.ObjectId `bson:"app,omitempty"`
	Name string `bson:"name"`
	Value string `bson:"value"`
}
