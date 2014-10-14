package nodegear

import (
	"gopkg.in/mgo.v2/bson"
	"../connection"
	"../models"
	"time"
)

func (p *Instance) createUptime() models.AppProcessUptime {
	uptime := &models.AppProcessUptime{
		ID: bson.NewObjectId(),
		Created: time.Now(),
		App: p.App_id,
		Process: p.Process_id,
		Server: Server.ID,
		Price_per_hour: Server.Price_per_hour,
		User: p.GetAppModel(&bson.M{ "user": 1 }).User,
	}

	c := connection.MongoC(models.AppProcessUptimeC)
	if err := c.Insert(&uptime); err != nil {
		panic(err)
	}

	p.Uptime = uptime.ID

	return *uptime
}

func (p *Instance) GetUptime() models.AppProcessUptime {
	if p.Uptime.Hex() == "" {
		return p.createUptime()
	}

	uptime := p.FindUptime(p.Uptime)

	if uptime.ID.Hex() == "" {
		return p.createUptime()
	}

	return uptime
}

func (p *Instance) FindUptime(id bson.ObjectId) models.AppProcessUptime {
	c := connection.MongoC(models.AppProcessUptimeC)
	var uptime models.AppProcessUptime

	if err := c.FindId(id).One(&uptime); err != nil {
		panic(err)
	}

	return uptime
}

func (p *Instance) RestartUptime() {
	uptime := p.createUptime()
	uptime.Start = time.Now()

	c := connection.MongoC(models.AppProcessUptimeC)
	if err := c.UpdateId(uptime.ID, uptime); err != nil {
		panic(err)
	}
}
