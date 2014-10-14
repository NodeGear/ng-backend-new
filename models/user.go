package models

import "gopkg.in/mgo.v2/bson"
import "time"

var UserC string = "users"

type User struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	
	Created time.Time `bson:"created"`
	Invitation_complete bool `bson:"invitation_complete"`

	Username string `bson:"username"`
	UsernameLowercase string `bson:"usernameLowercase"`
	
	Name string `bson:"name"`
	Email string `bson:"email"`
	Email_verified bool `bson:"email_verified"`
	
	Password string `bson:"password"`
	Is_new_pwd bool `bson:"is_new_pwd"`
	UpdatePassword bool `bson:"updatePassword"`
	
	Admin bool `bson:"admin"`
	
	Disabled bool `bson:"disabled"`
	
	Balance float32 `bson:"balance"`

	Stripe_customer string `bson:"stripe_customer"`
	Default_payment_method bson.ObjectId `bson:"default_payment_method,omitempty"`
	
	Tfa_enabled bool `bson:"tfa_enabled"`
	Tfa bson.ObjectId `bson:"tfa,omitempty"`
	
	AppLimit int `bson:"appLimit"`
	Newsletter_active bool `bson:"newsletter_active"`
}
