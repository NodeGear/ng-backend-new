package models

import "gopkg.in/mgo.v2/bson"
import "time"

var EmailVerificationC string = "emailverifications"

type EmailVerification struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	
	Created time.Time `bson:"created"`
	User bson.ObjectId `bson:"user,omitempty"`
	Email string `bson:"email"`
	Code string `bson:"code"`
	Verified bool `bson:"verified"`
	VerifiedDate time.Time `bson:"verifiedDate"`
}
