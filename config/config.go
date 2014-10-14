package config

import (
	"gopkg.in/mgo.v2/bson"
	"time"
	"encoding/json"
	"os"
	"fmt"
	"strconv"
)

type smtpConfiguration struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

type ServerConfiguration struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`

	Created time.Time `json:"created" bson:"created"`
	
	Name string `json:"name" bson:"name"`
	Location string `json:"location" bson:"location"`
	Identifier string `json:"identifier" bson:"identifier"`
	Price_per_hour float32 `json:"price_per_hour" bson:"price_per_hour"`
	
	App_memory int `json:"app_memory" bson:"app_memory"`
	AppLimit int `json:"appLimit" bson:"appLimit"`
	AppsRunning int `json:"appsRunning" bson:"appsRunning"`
}

type storageConfiguration struct {
	Enabled bool `json:"enabled"`
	Server string `json:"server"`
	Auth string `json:"auth"`
}

type JSONConfiguration struct {
	Redis_port int `json:"redis_port"`
	Redis_host string `json:"redis_host"`
	Redis_key string `json:"redis_key"`
	
	Bugsnag_key string `json:"bugsnag_key"`
	
	Smtp smtpConfiguration `json:"smtp"`
	
	Db string `json:"db"`
	
	Server ServerConfiguration `json:"server"`
	Storage storageConfiguration `json:"storage"`
	
	Statsd_ip string `json:"statsd_ip"`
	Statsd_port int16 `json:"statsd_port"`

	Docker_socket bool `json:"docker_socket"`
	Docker_ip string `json:"docker_ip"`
	Docker_port int `json:"docker_port"`

	Docker_url string

	Homepath string `json:"homepath"`
	Scriptspath string `json:"scriptspath"`
}

var Configuration JSONConfiguration

func init () {
	path := os.Getenv("CONFIG_PATH")
	if &path == nil || len(path) == 0 {
		path = "configuration.json"
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Println(path + " missing")
		panic(err)
	}

	configuration := JSONConfiguration{}

	if err := json.NewDecoder(file).Decode(&configuration); err != nil {
		fmt.Println("Could not decode configuration! (you're a moron)")
		panic(err)
	}

	Configuration = configuration

	Configuration.Docker_url = "http://" + Configuration.Docker_ip + ":" + strconv.Itoa(Configuration.Docker_port)
	
	if Configuration.Docker_socket {
		Configuration.Docker_url = "http://127.0.0.1:" + strconv.Itoa(Configuration.Docker_port)
	}

	fmt.Printf("Server: %v, IP: %v\n\n", Configuration.Server.Identifier, Configuration.Server.Location)
}
