package models

import "gopkg.in/mgo.v2/bson"
import "time"

var ForgotVerificationC string = "forgotverifications"

type ForgotVerification struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	
	Created time.Time `bson:"created"`
	User bson.ObjectId `bson:"user,omitempty"`
	Email string `bson:"email"`
	Code string `bson:"code"`
	Used bool `bson:"used"`
	UsedDate time.Time `bson:"usedDate"`
}
