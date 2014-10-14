package models

import "gopkg.in/mgo.v2/bson"
import "time"

var DatabaseC string = "databases"

type Database struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	
	Created time.Time `bson:"created"`
	Deleted bool `bson:"deleted"`
	
	User bson.ObjectId `bson:"user,omitempty"`
	
	Name string `bson:"name"`
	NameLowercase string `bson:"nameLowercase"`
	
	Database_type string `bson:"database_type"`
	
	Db_host string `bson:"db_host"`
	Db_user string `bson:"db_user"`
	Db_pass string `bson:"db_pass"`
	Db_name string `bson:"db_name"`
	Db_port string `bson:"db_port"`
}
