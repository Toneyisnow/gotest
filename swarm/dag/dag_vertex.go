package dag

import (
	"errors"
	"github.com/golang/protobuf/ptypes"
	"github.com/smartswarm/go/log"
	"sync"
	"time"
)

var _generateVertexMutex sync.Mutex

func GenerateNewVertex(selfNode *DagNode, peerParent *DagVertex) (vertex *DagVertex, err error) {

	if selfNode == nil {
		log.W("GenerateNewVertex: selfNode is nil.")
		return nil, errors.New("GenerateNewVertex: selfNode is nil.")
	}

	_generateVertexMutex.Lock()
	defer _generateVertexMutex.Unlock()

	vertex = new(DagVertex)

	dagStorage := DagStorageGetInstance()
	vertexDataList := dagStorage.GetPendingPayloadData()

	vertex.CreatorNodeId = selfNode.NodeId
	content := new(DagVertexContent)
	content.TimeStamp, _ = ptypes.TimestampProto(time.Now())

	// Mutex to read from storage
	lastSelfVertexHash, _, _ := dagStorage.GetLastVertexOnNode(selfNode, true)
	content.SelfParentHash = lastSelfVertexHash

	if peerParent != nil {
		content.PeerParentHash = peerParent.Hash
	} else {
		content.PeerParentHash = ""
	}

	for _, d := range vertexDataList {
		content.Data = append(content.Data, []byte(d))
	}
	vertex.Content = content

	// Save to database
	vertex.SaveToStorage()

	err = nil
	return
}

func GenerateGeneticVertex(node *DagNode) *DagNode {


	return nil
}

func FindPossibleUnknownVertexesForNode(node *DagNode) (resultList []*DagVertex, err error) {

	resultList = make([]*DagVertex, 0)

	resultList = append(resultList, nil)

	err = nil
	return
}

func (this *DagVertex) SaveToStorage() (err error) {

	if this == nil {
		return
	}

	dagStorage := DagStorageGetInstance()
	err = dagStorage._storage.SaveProto("V:" + string(this.Hash), this)
	if err == nil {
		log.I("DagVertex saved to storage.")
	}

	return
}

