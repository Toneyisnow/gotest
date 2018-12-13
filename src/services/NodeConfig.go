package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type NodeConfig struct {

	Self *NodeInfo `json:"self"`
	NetworkPeers []NodeInfo  `json:"network_peers"`
}

type NodeInfo struct {

	Identifier string `json:"id"`
	PublickKey string `json:"public_key"`
	HostUrl string `json:"host_url"`			// The Url Host of the node
	ServerPort int `json:"server_port"`
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

func (this *NodeConfig) GetPeerById(nodeId string) *NodeInfo {

	return &this.NetworkPeers[0]
}
