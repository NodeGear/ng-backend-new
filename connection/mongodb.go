package connection

import (
	"gopkg.in/mgo.v2"
	"../config"
)

var mongodb_c *mgo.Session
var mongodb_db *mgo.Database

func Mongo() *mgo.Session {
	if mongodb_c == nil {
		mongo, err := mgo.Dial(config.Configuration.Db)

		if err != nil {
			panic(err)
		}

		mongodb_c = mongo
	}

	return mongodb_c
}

func MongoDb() *mgo.Database {
	if mongodb_db == nil {
		m := Mongo()
		mongodb_db = m.DB("ng")
	}

	return mongodb_db
}

func MongoC (collection string) *mgo.Collection {
	return MongoDb().C(collection)
}
