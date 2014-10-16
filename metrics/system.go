package metrics

import (
	"gopkg.in/mgo.v2/bson"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
	"../connection"
	"../models"
	"../nodegear"
)

type stat struct {
	LastContainers int

	// new - old info..
	Total uint64 `json:"total"`
	Idle uint64 `json:"idle"`
	User uint64 `json:"user"`
	Sys uint64 `json:"sys"`

	// information t-1 second
	TotalOld uint64
	IdleOld uint64
	UserOld uint64
	SysOld uint64

	// CPU Free
	Free float64 `json:"free"`
	// Memory Free
	Mem float64 `json:"mem"`
	
	// Number of processors
	Processors int `json:"cores"`
	// Total memory (bytes)
	MemTotal uint64 `json:"memTotal"`
	// Total free memory (bytes)
	MemFree uint64 `json:"memFree"`
	// Server ID
	Identifier string `json:"identifier"`
	// Number of containers managed
	Containers int `json:"processes"`

	// Boot time
	BootTime time.Time `json:"btime"`
	// Number of processes total (on the OS)
	Processes uint64 `json:"procs"`
	// Number of running processes (on the OS)
	ProcsRunning uint64 `json:"procs_running"`
	// Number of blocked processes (on the OS)
	ProcsBlocked uint64 `json:"procs_blocked"`
}

func SystemStats() {
	system := &stat{}

	system.Processors = getNumProcessors()

	system.Identifier = nodegear.Server.Identifier

	ticker := time.NewTicker(time.Second)
	for _ = range ticker.C {
		memInfo(system)
		cpuInfo(system)

		numContainers := len(*nodegear.Instances)
		if system.LastContainers != numContainers {
			system.LastContainers = numContainers

			c := connection.MongoC(models.ServerC)
			err := c.UpdateId(nodegear.Server.ID, &bson.M{
				"$set": &bson.M{
					"appsRunning": numContainers,
				},
			})

			if err != nil {
				panic(err)
			}
		}

		system.Containers = numContainers

		redis := connection.Redis().Get()
		enc, err := json.Marshal(system)
		if err != nil {
			panic(err)
		}

		redis.Do("PUBLISH", "server_stats", enc)
		redis.Close()
	}
}

func memInfo(system *stat) {
	b, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(b), "\n")

	for _, line := range lines {
		fields := strings.SplitN(line, ":", 2)
		if len(fields) < 2 {
			continue
		}

		keyField := fields[0]

		if !(keyField == "MemFree" || keyField == "MemTotal") {
			continue
		}

		valFields := strings.Fields(fields[1])
		val, _ := strconv.ParseUint(valFields[0], 10, 64)
		
		if keyField == "MemFree" {
			system.MemFree = val * 1024
		} else if keyField == "MemTotal" {
			system.MemTotal = val * 1024
		}
	}

	system.Mem = float64(system.MemFree) / float64(system.MemTotal)
}

func cpuInfo(system *stat) {
	b, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(b), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)

		if len(fields) == 0 {
			continue
		}
		
		if fields[0] == "cpu" {
			old_idle := system.IdleOld
			old_sys := system.SysOld
			old_user := system.UserOld
			old_total := system.TotalOld

			system.UserOld, _ = strconv.ParseUint(fields[1], 10, 32)
			system.SysOld, _ = strconv.ParseUint(fields[3], 10, 32)
			system.IdleOld, _ = strconv.ParseUint(fields[4], 10, 32)
			nice, _ := strconv.ParseUint(fields[2], 10, 32)
			irq, _ := strconv.ParseUint(fields[6], 10, 32)
			system.TotalOld = system.IdleOld + system.UserOld + system.SysOld + irq + nice
			
			system.Idle = system.IdleOld - old_idle
			system.Sys = system.SysOld - old_sys
			system.User = system.UserOld - old_user
			system.Total = system.TotalOld - old_total

			system.Free = float64(system.Idle) / float64(system.Total)
		} else if fields[0] == "btime" {
			seconds, _ := strconv.ParseInt(fields[1], 10, 64)
			system.BootTime = time.Unix(seconds, 0)
		} else if fields[0] == "processes" {
			system.Processes, _ = strconv.ParseUint(fields[1], 10, 64)
		} else if fields[0] == "procs_running" {
			system.ProcsRunning, _ = strconv.ParseUint(fields[1], 10, 64)
		} else if fields[0] == "procs_blocked" {
			system.ProcsBlocked, _ = strconv.ParseUint(fields[1], 10, 64)
		}
	}
}

func getNumProcessors() int {
	var procs int

	// Get number of processors
	b, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(b), "\n")
	for i, line := range lines {
		if len(line) == 0 && i != len(lines)-1 {
			procs++
		}
	}

	return procs
}
