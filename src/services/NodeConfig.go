package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type NodeConfig struct {

	HostUrl string `json:"HostUrl"`			// The Url Host of the node
	ServerPort int `json:"server_port"`
	Identifier string `json:"Identifier"`
	PublickKey string `json:"PublicKey"`
}

func LoadConfigFromFile() *NodeConfig {

	jsonFile, err := os.Open("node-config.json")

	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	mainConfig := new(NodeConfig)
	json.Unmarshal(byteValue, mainConfig)

	return mainConfig
}