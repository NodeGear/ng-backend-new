package models

import "gopkg.in/mgo.v2/bson"

var AnalyticC string = "analytics"

type Analytic struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	
	App bson.ObjectId `bson:"app"`
	Process bson.ObjectId `bson:"process"`
	Hostname string `bson:"hostname"`
	Found bool `bson:"found"`
	Method string `bson:"method"`
	Url string `bson:"url"`
	StatusCode int `bson:"statusCode"`
	ReqSize int `bson:"reqSize"`
	ResSize int `bson:"resSize"`
	Ip string `bson:"ip"`
	Websocket bool `bson:"websocket"`
	Error bool `bson:"error"`
	ErrorCode string `bson:"errorCode"`
	Errno string `bson:"Errno"`
}
