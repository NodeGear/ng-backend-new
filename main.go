package main

import (
	"./routes"
	"./connection"
	"./nodegear"
	"./metrics"
	"fmt"
)

func main() {
	// Establish mongodb connection
	connection.Mongo()
	// Get redis conn
	connection.Redis()

	nodegear.Init()
	go metrics.SystemStats()
	go metrics.ContainerStats()

	fmt.Println("Fetching Processes")
	nodegear.FetchPreviousInstances()

	go nodegear.ListenForEvents()
	routes.ListenToRedis()

	fmt.Println("Hello!")
}