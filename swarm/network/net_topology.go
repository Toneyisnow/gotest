package network

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type NetTopology struct {

	_self *NetDevice	 `json:"self"`
	_peers []*NetDevice  `json:"peers"`
}

func LoadTopology() *NetTopology {

	return LoadTopologySampleTest()
	//// return LoadTopologyFromJsonFile("net-topology.json")
}

func LoadTopologyFromDB(dbName string) *NetTopology {

	// Load the topology information from database
	return LoadTopologySampleTest()

}

func LoadTopologyFromJsonFile(jsonFileName string) *NetTopology {

	jsonFile, err := os.Open(jsonFileName)

	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	topology := new(NetTopology)
	json.Unmarshal(byteValue, topology)

	return topology
}

func LoadTopologySampleTest() *NetTopology {

	// Load Test data




	topology := new(NetTopology)

	topology._self = CreateDevice("127.0.0.1", 8881)

	peer1 := CreateDevice("127.0.0.1", 8881)
	peer2 := CreateDevice("127.0.0.1", 8882)
	topology._peers = make([]*NetDevice, 0)
	topology._peers = append(topology._peers, peer1, peer2)

	return topology
}

func (this *NetTopology) GetPeerDeviceByAddress(address string) *NetDevice {

	if (len(address) == 0) {
		return nil
	}

	for _, d := range this._peers {

		// Here we are not sure HostUrl/IP which one will be used in testing phase, will update later
		if (strings.Compare(d.HostUrl, address) == 0 || strings.Compare(d.IPAddress, address) == 0) {
			return d
		}
	}

	return nil
}

func (this *NetTopology) Self() *NetDevice {
	return this._self
}
