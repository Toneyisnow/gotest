package dag

import (
	"../network"
	"encoding/json"
	"fmt"
	"github.com/smartswarm/go/log"
	"io/ioutil"
	"os"
)

type DagNode struct {

	NodeId uint64
	Device *network.NetDevice

	GenesisVertexHash []byte

}

type DagNodes struct {

	Self *DagNode
	Peers []*DagNode
}

func LoadDagNodesFromFile(jsonFileName string) *DagNodes {

	log.I("Begin LoadDagNodesFromFile")
	jsonFile, err := os.Open(jsonFileName)

	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)
	dagNodes := new(DagNodes)
	json.Unmarshal(byteValue, dagNodes)

	jsonFile.Close()

	log.I("End LoadDagNodesFromFile")

	/*
	pubKey, privKey := secp256k1.GenerateKeyPair()

	dagNodes.Self.Device.PublicKey = pubKey
	dagNodes.Self.Device.PrivateKey = privKey

	for _, n := range dagNodes.Peers {

		pubKey, privKey = secp256k1.GenerateKeyPair()
		n.Device.PublicKey = pubKey
	}

	dagNodes.SaveToFile(jsonFileName)
	*/

	return dagNodes
}

func (this *DagNodes) SaveToFile(jsonFileName string) (err error) {

	log.I("Begin SaveToFile")

	contentBytes, err := json.MarshalIndent(this, "", "  ")
	if err != nil {
		fmt.Println(err)
		return err
	}

	ioutil.WriteFile(jsonFileName, contentBytes, 0644)

	log.I("End SaveToFile")

	return nil
}

func (this *DagNodes) GetSelf() (*DagNode) {
	return this.Self
}

func (this *DagNodes) AllNodes() []*DagNode {

	return append(this.Peers, this.Self)
}

func (this *DagNodes) GetPeerNodeById(nodeId uint64) (*DagNode) {

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

func (this *DagNodes) GetMajorityCount() int {

	return len(this.AllNodes()) * 2 / 3
}