package main

import (
	"fmt"
	"node"
	"os"
	"io/ioutil"
	"encoding/json"
	"strconv"
)

// User struct which contains a name
// a type and a list of social links
type Config struct {
	ServerPort   int `json:"server_port"`
	PeerPort     int `json:"peer_port"`
}

func main() {
	
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
	
	
	e := node.Engine { ServerPort: mainConfig.ServerPort, PeerPort: mainConfig.PeerPort }
	
	e.Start("abcde")
	
	
	fmt.Println("Exit.")
}