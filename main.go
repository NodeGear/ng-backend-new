package main

import (
	"./routes"
	"./connection"
	"./nodegear"
	"fmt"
)

func main() {
	// Establish mongodb connection
	connection.Mongo()
	// Get redis conn
	connection.Redis()

	nodegear.Init()

	fmt.Println("Fetching Processes")
	nodegear.FetchPreviousInstances()

	go nodegear.ListenForEvents()
	routes.ListenToRedis()

	fmt.Println("Hello!")
}