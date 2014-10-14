package models

import "gopkg.in/mgo.v2/bson"
import "time"

var AppProcessUptimeC string = "appprocessuptimes"

type AppProcessUptime struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	
	Created time.Time `bson:"created"`
	User bson.ObjectId `bson:"user,omitempty"`
	App bson.ObjectId `bson:"app,omitempty"`
	Process bson.ObjectId `bson:"process,omitempty"`
	Server bson.ObjectId `bson:"server,omitempty"`
	Minutes int `bson:"minutes"`
	Start time.Time `bson:"start"`
	End time.Time `bson:"end"`
	Sealed bool `bson:"sealed"`
	Price_per_hour float32 `bson:"price_per_hour"`
	Paid bool `bson:"paid"`
}
