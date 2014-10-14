package nodegear

import (
	"testing"
	"time"
	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2/bson"
	"../config"
	"../connection"
	"../models"
	"."
	"fmt"
	"encoding/json"
)

var ProxyTest *proxyTest
type proxyTest struct {
	User bson.ObjectId
	App bson.ObjectId
	Process bson.ObjectId
}

type proxyDetails struct {
	Ssl bool `redis:"ssl"`
	Ssl_only bool `redis:"ssl_only"`
	Owner string `redis:"owner"`
}

func (p *proxyTest) setup(identifier string) (nodegear.Instance, []models.AppDomain) {
	nodegear.Init()

	if _, err := connection.MongoC(models.AppDomainC).RemoveAll(&bson.M{}); err != nil {
		panic(err)
	}
	if _, err := connection.MongoC(models.AppProcessC).RemoveAll(&bson.M{}); err != nil {
		panic(err)
	}

	p.User = bson.NewObjectId()
	p.App = bson.NewObjectId()
	p.Process = bson.NewObjectId()

	connection.MongoC(models.AppProcessC).Insert(&models.AppProcess{
		ID: p.Process,
	})

	instance := nodegear.Instance{
		App_id: p.App,
		Process_id: p.Process,
	}
	instance.Init()

	// Add a few domains to db
	var domains []models.AppDomain

	domains = append(domains, models.AppDomain{
		ID: bson.NewObjectId(),
		Created: time.Now(),
		App: p.App,
		User: p.User,
		Domain: "hello.ng-proxy" + identifier,
		Ssl: false,
		Ssl_only: false,
		Certificate: "",
		Certificate_key: "",
	}, models.AppDomain{
		ID: bson.NewObjectId(),
		Created: time.Now(),
		App: p.App,
		User: p.User,
		Domain: "hello-2.ng-proxy" + identifier,
		Ssl: true,
		Ssl_only: true,
		Certificate: "ad",
		Certificate_key: "adbc",
	})

	redis_c := connection.Redis()
	redis_c.Send("MULTI")

	for _, domain := range domains {
		redis_c.Send("DEL", "proxy:domain_details_" + domain.Domain)
		redis_c.Send("DEL", "proxy:domain_members_" + domain.Domain)

		if err := connection.MongoC(models.AppDomainC).Insert(&domain); err != nil {
			panic(err)
		}
	}

	redis_c.Send("EXEC")
	instance.AddToProxy()

	return instance, domains
}

func TestAddToProxy(t *testing.T) {
	ProxyTest = &proxyTest{}
	instance, domains := ProxyTest.setup("1")

	redis_c := connection.Redis()

	for _, domain := range domains {
		details_vars, err := redis.Values(redis_c.Do("HGETALL", "proxy:domain_details_" + domain.Domain))
		if err != nil {
			panic(err)
		}

		var details proxyDetails
		if err := redis.ScanStruct(details_vars, &details); err != nil {
			panic(err)
		}

		members, err := redis.Strings(redis_c.Do("SMEMBERS", "proxy:domain_members_" + domain.Domain))
		if err != nil {
			panic(err)
		}

		if details.Ssl != domain.Ssl || details.Ssl_only != domain.Ssl_only || details.Owner != domain.User.Hex() {
			fmt.Println(details, domain)
			t.Error("Details mismatch")
		}

		if len(members) != 1 {
			fmt.Println(members)
			t.Error("More than one member")
		}

		var member nodegear.ProxyMember
		if err := json.Unmarshal([]byte(members[0]), &member); err != nil {
			panic(err)
		}

		if member.Extra != instance.Process_id.Hex() || member.Owner != domain.User.Hex() || member.Hostname != config.Configuration.Server.Location || member.Port != instance.Port {
			fmt.Println(member, domain)
			t.Error("Details mismatch member vs domain")
		}
	}
}

func TestRemoveFromProxy(t *testing.T) {
	ProxyTest = &proxyTest{}
	instance, domains := ProxyTest.setup("2")

	redis_c := connection.Redis()

	instance.RemoveFromProxy()

	for _, domain := range domains {
		details, err := redis.Values(redis_c.Do("HGETALL", "proxy:domain_details_" + domain.Domain))
		if err != nil {
			panic(err)
		}

		members, err := redis.Strings(redis_c.Do("SMEMBERS", "proxy:domain_members_" + domain.Domain))
		if err != nil {
			panic(err)
		}

		if len(details) != 0 {
			fmt.Println(details, domain)
			t.Error("Details not empty mismatch")
		}

		if len(members) != 0 {
			fmt.Println(members)
			t.Error("More than 0 member")
		}
	}
}
