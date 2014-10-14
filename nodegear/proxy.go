package nodegear

import (
	"gopkg.in/mgo.v2/bson"
	"github.com/garyburd/redigo/redis"
	"../connection"
	"../models"
	"../config"
	"encoding/json"
)

type ProxyMember struct {
	Extra string `json:"extra"`
	Owner string `json:"owner"`
	Hostname string `json:"hostname"`
	Port int `json:"port"`
}

func getDomains(p *Instance) []models.AppDomain {
	c := connection.MongoC(models.AppDomainC)

	var domains []models.AppDomain
	
	c.Find(&bson.M{
		"app": p.App_id,
	}).All(&domains)

	return domains
}

func (p *Instance) AddToProxy() {
	domains := getDomains(p)

	if len(domains) == 0 {
		p.Log("\nNo Domains Set, your app won't be accessible! Halting startup.\n")
		(&models.AppEvent{
			App: p.App_id,
			Process: p.Process_id,
			Name: "Start Error",
			Message: "App could not be started because it has no defined domains.",
		}).Add()

		p.Stop()
		return
	}

	redis_connection := connection.Redis().Get()
	defer redis_connection.Close()

	redis_connection.Send("MULTI")
	for _, domain := range domains {
		redis_connection.Send("HMSET", "proxy:domain_details_" + domain.Domain, "ssl", domain.Ssl, "ssl_only", domain.Ssl_only, "owner", domain.User.Hex())
		
		member, _ := json.Marshal(&ProxyMember{
			Extra: p.Process_id.Hex(),
			Owner: domain.User.Hex(),
			Hostname: config.Configuration.Server.Location,
			Port: p.Port,
		})
		redis_connection.Send("SADD", "proxy:domain_members_" + domain.Domain, string(member))
	}

	redis_connection.Send("EXEC")
}

func (p *Instance) RemoveFromProxy() {
	domains := getDomains(p)
	redis_connection := connection.Redis().Get()
	defer redis_connection.Close()

	for _, domain := range domains {
		key := "proxy:domain_members_" + domain.Domain

		members_encoded, err := redis.Strings(redis_connection.Do("SMEMBERS", key))
		if err != nil {
			panic(err)
		}

		var members []ProxyMember
		for _, member_encoded := range members_encoded {
			var member ProxyMember
			if err := json.Unmarshal([]byte(member_encoded), &member); err != nil {
				panic(err)
			}

			members = append(members, member)
		}

		members_num := len(members)

		redis_connection.Send("MULTI")

		for _, member := range members {
			if member.Owner != domain.User.Hex() || member.Extra == p.Process_id.Hex() {
				encoded_member, err := json.Marshal(&member)
				if err != nil {
					panic(err)
				}

				redis_connection.Send("SREM", key, encoded_member)
				members_num--
			}
		}

		if members_num == 0 {
			redis_connection.Send("DEL", key)
			redis_connection.Send("DEL", "proxy:domain_details_" + domain.Domain)
		}

		redis_connection.Send("EXEC")
	}
}
