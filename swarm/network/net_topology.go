package network

import "strings"

type NetTopology struct {

	_self *NetDevice
	_peers []*NetDevice
}

func LoadTopologyFromDB(dbName string) *NetTopology {

	// Load the topology information from database

	return LoadTopologyFromJson("")
}

func LoadTopologyFromJson(jsonString string) *NetTopology {

	// Load Test data

	topology := new(NetTopology)

	topology._self = CreateDevice("127.0.0.1", 8883)

	peer1 := CreateDevice("127.0.0.1", 8881)
	peer2 := CreateDevice("127.0.0.1", 8882)
	topology._peers = make([]*NetDevice, 0)
	topology._peers = append(topology._peers, peer1, peer2)

	return topology
}

func (this *NetTopology) GetDeviceByAddress(address string) *NetDevice {

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