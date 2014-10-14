package models

import "gopkg.in/mgo.v2/bson"
import "time"

var AppDomainC string = "appdomains"

type AppDomain struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`

	Created time.Time `bson:"created"`
	App bson.ObjectId `bson:"app,omitempty"`
	User bson.ObjectId `bson:"user,omitempty"`
	Domain string `bson:"domain"`
	Ssl bool `bson:"ssl"`
	Ssl_only bool `bson:"ssl_only"`
	Certificate string `bson:"certificate"`
	Certificate_key string `bson:"certificate_key"`
}
