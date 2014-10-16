package metrics

import (
	"io/ioutil"
	"strconv"
	"strings"
	"time"
	"encoding/json"
	"../nodegear"
	"../connection"
)

type processStat struct {
	ProcessId string `json:"_id"`
	AppId string `json:"app"`
	Details *processStatDetails `json:"monitor"`
}

type processStatDetails struct {
	Rss uint64 `json:"rss"`
	Rss_max uint64 `json:"rss_max"`
	Cpu_percent int `json:"cpu_percent"`
	Cpu_percent_max int `json:"cpu_percent_max"`
}

func ContainerStats() {
	processors := getNumProcessors()

	ticker := time.NewTicker(time.Second)
	for _ = range ticker.C {
		instances := *nodegear.Instances
		for _, instance := range instances {
			if instance.Running == false {
				continue
			}

			stat := getLines("/sys/fs/cgroup/cpuacct/docker/" + instance.Container_id + "/cpuacct.stat")
			memstat := getLines("/sys/fs/cgroup/memory/docker/" + instance.Container_id + "/memory.stat")
			
			var user uint64
			var system uint64
			var rss uint64

			// Something's wrong
			if len(stat) == 0 || len(memstat) == 0 {
				continue
			}

			// Parse CPU usage
			user, _ = strconv.ParseUint(strings.Fields(stat[0])[1], 10, 32)
			system, _ = strconv.ParseUint(strings.Fields(stat[1])[1], 10, 32)
			total := user + system
			if instance.LastCPU == 0 {
				instance.LastCPU = total
			}

			// Diff between now and t-1 second ago
			cpu_perc := int(total - instance.LastCPU)
			instance.LastCPU = total

			// Parse memory usage
			for _, line := range memstat {
				fields := strings.Fields(line)
				if fields[0] != "total_rss" {
					continue
				}

				rss, _ = strconv.ParseUint(fields[1], 10, 64)
				if rss != 0 {
					rss = rss / 1024 // bytes to kb
				}
				break
			}

			s := &processStat{
				ProcessId: instance.Process_id.Hex(),
				AppId: instance.App_id.Hex(),
				Details: &processStatDetails{
					Rss: rss,
					Rss_max: instance.AppMemory / 1024,
					Cpu_percent: cpu_perc,
					Cpu_percent_max: 100 * processors,
				},
			}

			monitor, err := json.Marshal(s)
			if err != nil {
				panic(err)
			}
			
			redis := connection.Redis().Get()
			redis.Do("PUBLISH", "process_stats", monitor)
			redis.Close()
		}
	}
}

func getLines(path string) []string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return make([]string, 0)
	}

	return strings.Split(string(b), "\n")
}
