package main

import (
	"fmt"
	"time"
	"services"
)


func main() {

	server := new(services.NodeServer)

	server.Initialize()

	fmt.Println("Server Starting...")
	server.Start()
	fmt.Println("Server Started.")

	time.Sleep(10 * time.Second)

	fmt.Println("Server Stopping...")
	server.Stop()
	fmt.Println("Server Stopped.")

	/*
	jsonString := `{ "HostUrl":"localhost:9999", "Identifier":"1" }`
	
	node := objectmodels.ReadNodeFromJson(jsonString)
	fmt.Println("Node: " + node.HostUrl)


	jsonFile, err := os.Open("node-config.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var mainConfig Config
	json.Unmarshal(byteValue, &mainConfig)

	fmt.Println("Server Port: " + strconv.Itoa(mainConfig.ServerPort))
	
	
	e := objectmodels.Engine { ServerPort: mainConfig.ServerPort, PeerPort: mainConfig.PeerPort }
	
	e.Start("abcde")
	*/
	
	fmt.Println("Exit.")
}