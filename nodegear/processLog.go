package nodegear

import (
	//"fmt"
	"../connection"
)

func (p *Instance) Log(what string) {
	//fmt.Printf("%s", what)

	redis := connection.Redis().Get()
	defer redis.Close()

	redis.Send("MULTI")

	if p.Inserted_log_to_redis == false {
		p.Inserted_log_to_redis = true
		redis.Send("LPUSH", "pm:app_process_logs_" + p.Process_id.Hex(), p.CurrentLog)
		redis.Send("PUBLISH", "pm:app_log_new", p.Process_id.Hex())
	}

	redis.Send("LPUSH", "pm:app_process_log_" + p.CurrentLog, what)
	redis.Send("PUBLISH", "pm:app_log_entry", p.Process_id.Hex() + "|" + what)

	if _, err := redis.Do("EXEC"); err != nil {
		panic(err)
	}
}
