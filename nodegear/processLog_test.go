package nodegear

import (
	"testing"
	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2/bson"
	"../connection"
	"."
	"fmt"
)

var ProcessLogTest processLogTest
type processLogTest struct {
	Process bson.ObjectId
}

func (p processLogTest) setup() nodegear.Instance {
	nodegear.Init()

	p.Process = bson.NewObjectId()

	instance := nodegear.Instance{
		Process_id: p.Process,
		CurrentLog: p.Process.Hex() + "-log",
	}

	return instance
}

func TestProcessLog(t *testing.T) {
	instance := ProcessLogTest.setup()

	if instance.Inserted_log_to_redis == true {
		t.Error("Inserted_log_to_redis is true")
	}

	instance.Log("Hello")

	if instance.Inserted_log_to_redis == false {
		t.Error("Log did not insert log to redis")
	}

	redis_c := connection.Redis()
	length, err := redis.Int(redis_c.Do("LLEN", "pm:app_process_logs_" + instance.Process_id.Hex()))

	if length != 1 || err != nil {
		t.Error("Did not insert log", err)
	}

	length, err = redis.Int(redis_c.Do("LLEN", "pm:app_process_log_" + instance.CurrentLog))

	if length != 1 || err != nil {
		t.Error("Did not insert log message", err)
	}

	logs, err := redis.Strings(redis_c.Do("LRANGE", "pm:app_process_log_" + instance.CurrentLog, 0, 1))
	if err != nil {
		panic(err)
	}

	if logs[0] != "Hello" {
		fmt.Println(logs)
		t.Error("Wrong log message")
	}
}
