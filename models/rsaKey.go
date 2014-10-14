package models

import "gopkg.in/mgo.v2/bson"
import "time"

var RSAKeyC string = "rsakeys"

type RSAKey struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	
	Created time.Time `bson:"created"`
	Deleted bool `bson:"deleted"`
	User bson.ObjectId `bson:"user,omitempty"`
	Name string `bson:"name"`
	nameLowercase string `bson:"nameLowercase"`
	System_key bool `bson:"system_key"`
	Private_key string `bson:"private_key"`
	Public_key string `bson:"public_key"`
	Installed bool `bson:"installed"`
	Installing bool `bson:"installing"`
}