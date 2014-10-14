package models

import "gopkg.in/mgo.v2/bson"
import "time"

var AppC string = "apps"

type App struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	
	Created time.Time `bson:"created"`
	App_type string `bson:"app_type"` //default 'node'
	Name string `bson:"name"`
	NameUrl string `bson:"nameUrl"`
	NameLowercase string `bson:"nameLowercase,omitempty"`
	User bson.ObjectId `bson:"user"`
	Deleted bool `bson:"deleted"`
	Location string `bson:"location"`
	Branch string `bson:"branch"` // default 'master'
	Docker struct {
		Image string `bson:"image"`
		Command string `bson:"command"`
		Links []struct {
			App bson.ObjectId `bson:"app,omitempty"`
			Name string `bson:"name"`
		} `bson:"links"`
	} `bson:"docker"`
}
