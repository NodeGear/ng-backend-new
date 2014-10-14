package nodegear

import (
	"testing"
	"time"
	"gopkg.in/mgo.v2/bson"
	"../connection"
	"../models"
	"."
)

var UptimeTest *uptimeTest
type uptimeTest struct {
	User models.User
	App models.App
	Process models.AppProcess
}

func (t *uptimeTest) setup() {
	nodegear.Init()

	if _, err := connection.MongoC(models.UserC).RemoveAll(&bson.M{}); err != nil {
		panic(err)
	}
	if _, err := connection.MongoC(models.AppC).RemoveAll(&bson.M{}); err != nil {
		panic(err)
	}
	if _, err := connection.MongoC(models.AppProcessC).RemoveAll(&bson.M{}); err != nil {
		panic(err)
	}

	t.User = models.User{
		ID: bson.NewObjectId(),
	}

	t.App = models.App{
		ID: bson.NewObjectId(),
		Created: time.Now(),
		App_type: "node",
		Name: "test",
		NameUrl: "test",
		NameLowercase: "test",
		User: t.User.ID,
		Deleted: false,
		Location: "git@github.com/nodegear/node-js-sample.git",
		Branch: "master",
		Docker: []models.App_docker{},
	}

	t.Process = models.AppProcess{
		ID: bson.NewObjectId(),
		Created: time.Now(),
		Name: "testProcess",
		App: t.App.ID,
		Server: nodegear.Server.ID,
		Running: false,
		Restarts: false,
		Deleted: false,
		ContainerID: "",
	}
	
	if err := connection.MongoC(models.UserC).Insert(&t.User); err != nil {
		panic(err)
	}
	if err := connection.MongoC(models.AppC).Insert(&t.App); err != nil {
		panic(err)
	}
	if err := connection.MongoC(models.AppProcessC).Insert(&t.Process); err != nil {
		panic(err)
	}
}

func TestSetup(t *testing.T) {
	UptimeTest = &uptimeTest{}
	UptimeTest.setup()
}

func TestCreateUptime(t *testing.T) {
	instance := nodegear.Instance{
		App_id: UptimeTest.App.ID,
		Process_id: UptimeTest.Process.ID,
	}

	uptime := instance.GetUptime()

	if instance.Uptime.Hex() == "" {
		t.Error("Did not Set Uptime on Instance")
	}

	// Repeated GetUptime should yield the same ID
	if up := instance.GetUptime(); up.ID.Hex() != uptime.ID.Hex() {
		t.Error("Running GetUptime twice does not yield same uptime")
	}
}

func TestFindUptime(t *testing.T) {
	uptimeModel := models.AppProcessUptime{
		ID: bson.NewObjectId(),
		Minutes: 2,
	}

	if err := connection.MongoC(models.AppProcessUptimeC).Insert(&uptimeModel); err != nil {
		panic(err)
	}

	uptime := (&nodegear.Instance{}).FindUptime(uptimeModel.ID)

	if uptime.ID.Hex() != uptimeModel.ID.Hex() {
		t.Error("Did not find uptime")
	}
}
