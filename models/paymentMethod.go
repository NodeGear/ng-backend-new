package models

import "gopkg.in/mgo.v2/bson"
import "time"

var PaymentMethodC string = "paymentmethods"

type PaymentMethod struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	
	Type string `bson:"type"`
	Id string `bson:"id"`
	Name string `bson:"name"`
	Cardholder string `bson:"cardholder"`
	Created time.Time `bson:"created"`
	Last4 string `bson:"last4"`
	Disabled bool `bson:"disabled"`
	User bson.ObjectId `bson:"user,omitempty"`
}
