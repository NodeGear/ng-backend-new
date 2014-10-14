package models

import "gopkg.in/mgo.v2/bson"
import "time"

var InvitationC string = "invitations"

type Invitation struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	
	Created time.Time `bson:"created"`
	User bson.ObjectId `bson:"user,omitempty"`
	IsConfirmed bool `bson:"isConfirmed"`
	Confirmed time.Time `bson:"confirmed"`
	Confirmed_by time.Time `bson:"confirmed_by"`
}
