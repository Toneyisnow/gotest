package dag

import (
	"crypto/sha256"
	"errors"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/smartswarm/core/crypto/secp256k1"
	"github.com/smartswarm/go/log"
	"sync"
	"time"
)

var _createVertexMutex sync.Mutex

func CreateVertex(dagStorage *DagStorage, selfNode *DagNode, peerParent *DagVertex) (vertex *DagVertex, err error) {

	if selfNode == nil {
		log.W("CreateVertex: selfNode is nil.")
		return nil, errors.New("CreateVertex: selfNode is nil.")
	}

	_createVertexMutex.Lock()
	defer _createVertexMutex.Unlock()

	vertex = new(DagVertex)

	vertex.CreatorNodeId = selfNode.NodeId
	content := new(DagVertexContent)
	content.TimeStamp, _ = ptypes.TimestampProto(time.Now())

	// Mutex to read from storage
	lastSelfVertexHash, _, _ := dagStorage.GetLastVertexOnNode(selfNode, true)
	content.SelfParentHash = lastSelfVertexHash

	if peerParent != nil {
		content.PeerParentHash = peerParent.Hash
	} else {
		content.PeerParentHash = nil
	}

	dagStorage.queuePendingData.StartIterate()
	for  {
		_, data := dagStorage.queuePendingData.IterateNext()
		if data == nil {
			break
		}
		content.Data = append(content.Data, []byte(data))
	}
	vertex.Content = content

	// Calculate Hash and Encrypt it
	contentBytes, err := proto.Marshal(content)
	if err != nil {
		return nil, errors.New("Fatal: proto Marshal content failed.")
	}

	sha256Hash := sha256.Sum256(contentBytes)
	vertex.Hash = sha256Hash[:]

	privateKey := selfNode.Device.PrivateKey
	signature, err := secp256k1.Sign(vertex.Hash, privateKey)
	vertex.Signature = signature

	// Save to database
	SaveVertex(dagStorage, vertex)

	// Clear the pending queue data
	dagStorage.queuePendingData.Clear()

	err = nil
	return
}

func GenerateGeneticVertex(node *DagNode) *DagNode {


	return nil
}


