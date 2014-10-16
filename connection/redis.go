package connection

import (
	"github.com/garyburd/redigo/redis"
	"../config"
	"strconv"
	"time"
)

var redis_c *redis.Pool

func Redis() *redis.Pool {
	if redis_c == nil {
		redis_c = &redis.Pool{
			MaxIdle: 3,
			IdleTimeout: 240 * time.Second,
			Dial: func () (redis.Conn, error) {
				c, err := redis.Dial("tcp", config.Configuration.Redis_host + ":" + strconv.Itoa(config.Configuration.Redis_port))
				if err != nil {
					return nil, err
				}

				if len(config.Configuration.Redis_key) > 0 {
					if _, err := c.Do("AUTH", config.Configuration.Redis_key); err != nil {
						c.Close()
						return nil, err
					}
				}
				
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		}
	}

	return redis_c
}
