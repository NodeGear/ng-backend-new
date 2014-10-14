package nodegear

import (
	"testing"
//	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2/bson"
	"../connection"
	"../models"
	"."
	"time"
)

func TestProcess(t *testing.T) {
	nodegear.Init()

	t.Error("lol")

	if _, err := connection.MongoC(models.UserC).RemoveAll(&bson.M{}); err != nil {
		panic(err)
	}
	if _, err := connection.MongoC(models.AppC).RemoveAll(&bson.M{}); err != nil {
		panic(err)
	}
	if _, err := connection.MongoC(models.AppProcessC).RemoveAll(&bson.M{}); err != nil {
		panic(err)
	}

	User := models.User{
		ID: bson.NewObjectId(),
	}

	App := models.App{
		ID: bson.NewObjectId(),
		Created: time.Now(),
		App_type: "node",
		Name: "test",
		NameUrl: "test",
		NameLowercase: "test",
		User: User.ID,
		Deleted: false,
		Location: "git://github.com/nodegear/node-js-sample.git",
		Branch: "master",
		Docker: []models.App_docker{},
	}

	Process := models.AppProcess{
		ID: bson.NewObjectId(),
		Created: time.Now(),
		Name: "testProcess",
		App: App.ID,
		Server: nodegear.Server.ID,
		Running: false,
		Restarts: false,
		Deleted: false,
		ContainerID: "",
	}

	RSAKey := models.RSAKey{
		ID: bson.NewObjectId(),
	}
	
	if err := connection.MongoC(models.UserC).Insert(&User); err != nil {
		panic(err)
	}
	if err := connection.MongoC(models.AppC).Insert(&App); err != nil {
		panic(err)
	}
	if err := connection.MongoC(models.AppProcessC).Insert(&Process); err != nil {
		panic(err)
	}
	if err := connection.MongoC(models.RSAKeyC).Insert(&RSAKey); err != nil {
		panic(err)
	}

	instance := nodegear.Instance{
		App_id: App.ID,
		Process_id: Process.ID,
	}
	instance.Init()
	nodegear.Instances = append(nodegear.Instances, &instance)

	go nodegear.ListenForEvents()

	time.Sleep(1 * time.Second)
	instance.Start()

	time.Sleep(2 * time.Second)
}