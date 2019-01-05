package dag

import (
	"../network"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type DagNode struct {

	NodeId int64
	Device *network.NetDevice

	GenesisVertexHash string
}

type DagNodes struct {

	Self *DagNode
	Peers []*DagNode
}

func LoadDagNodesFromFile(jsonFileName string) *DagNodes {

	jsonFile, err := os.Open(jsonFileName)

	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	dagNodes := new(DagNodes)
	json.Unmarshal(byteValue, dagNodes)

	return dagNodes
}

func (this *DagNodes) GetSelf() (*DagNode) {
	return this.Self
}

func (this *DagNodes) GetPeerNodeById(nodeId int64) (*DagNode) {

	if (nodeId == 0) {
		return nil
	}

	for _, d := range this.Peers {

		if (d.NodeId == nodeId) {
			return d
		}
	}

	return nil
}

func (this *DagNodes) GetNetTopology() *network.NetTopology {

	topology := new(network.NetTopology)

	topology.SetSelf(this.Self.Device)

	for _, node := range this.Peers {
		device := node.Device
		topology.AddPeer(device)
	}

	return topology
}

func (this *DagNodes) FindCandidateNodesToSendVertex(maxCount int) []*DagNode {

	result := make([]*DagNode, 0)



	return result
}