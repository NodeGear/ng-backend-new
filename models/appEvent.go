package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
	"../connection"
)

var AppEventsC string = "appevents"

type AppEvent struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	
	Created time.Time `bson:"created"`
	App bson.ObjectId `bson:"app,omitempty"`
	Process bson.ObjectId `bson:"process,omitempty"`
	Name string `bson:"name"`
	Message string `bson:"message"`
}

func (ev *AppEvent) Add () {
	c := connection.MongoC(AppEventsC)

	if len(ev.ID.Hex()) == 0 {
		ev.ID = bson.NewObjectId()
	}

	ev.Created = time.Now()
	if err := c.Insert(ev); err != nil {
		panic(err)
	}

	redis_c := connection.Redis().Get()
	defer redis_c.Close()
	
	if _, err := redis_c.Do("PUBLISH", "pm:app_event", ev.ID.Hex()); err != nil {
		panic(err)
	}

	return
}
