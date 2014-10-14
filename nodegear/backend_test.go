package nodegear

import (
	"testing"
	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2/bson"
	"../config"
	"../connection"
	"../models"
	"."
	"fmt"
)

func TestInit(t *testing.T) {
	c := connection.MongoC(models.ServerC)
	if err := c.Remove(&bson.M{
		"identifier": config.Configuration.Server.Identifier,
	}); err != nil && err.Error() != "not found" {
		panic(err)
	}

	redis_c := connection.Redis()
	port, err := redis.Int(redis_c.Do("GET", "backend_port:" + config.Configuration.Server.Identifier))
	if err != nil {
		if err.Error() == "redigo: nil returned" {
			port = 9000
		} else {
			panic(err)
		}
	}

	// Should create a new server
	nodegear.Init()

	if _, err := redis_c.Do("SET", "backend_port:" + config.Configuration.Server.Identifier, port); err != nil {
		panic(err)
	}

	var server models.Server
	if err := c.Find(&bson.M{ "identifier": config.Configuration.Server.Identifier }).One(&server); err != nil {
		panic(err)
	}

	if server.ID.Hex() == "" {
		fmt.Println(server)
		t.Error("Did not create a new server")
	}
}
