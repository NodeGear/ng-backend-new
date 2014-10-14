package models

import "gopkg.in/mgo.v2/bson"
import "time"

var AppProcessDataSnapshotC string = "appprocessdatasnapshots"

type AppProcessDataSnapshot struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	
	Created time.Time `bson:"created"`
	App bson.ObjectId `bson:"app,omitempty"`
	OriginProcess bson.ObjectId `bson:"originProcess,omitempty"`
	OriginServer bson.ObjectId `bson:"originServer,omitempty"`
	ContentSize int `bson:"contentSize"`
}
