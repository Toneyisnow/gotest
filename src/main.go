package main

import (
	"fmt"
	"services"
	"time"
)


func main() {

	services.ExecuteNode()
}

func StartServer() {

	server := new(services.NodeServer)

	server.Initialize()

	fmt.Println("Server Starting...")
	server.Start()
	fmt.Println("Server Started.")

	time.Sleep(10 * time.Second)

	fmt.Println("Server Stopping...")
	server.Stop()
	fmt.Println("Server Stopped.")
}

func StartClient() {

	client := new(services.NodeClient)
	client.Initialize()

	config := services.LoadConfigFromFile()
	client.Connect(&config.NetworkPeers[0])

	client.SendMessage("Good to see that")

	time.Sleep(time.Second)
	client.SendMessage("Good to see that")

	time.Sleep(time.Second)
	client.SendMessage("Good to see that")

	time.Sleep(time.Second)
	client.SendMessage("Good to see that")

}