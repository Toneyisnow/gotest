package dag

import (
	"crypto/sha256"
	"errors"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/smartswarm/core/crypto/secp256k1"
	"github.com/smartswarm/go/log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var createVertexMutex sync.Mutex


func CreateSelfDataVertex(dagStorge *DagStorage, selfNode *DagNode, data PayloadData) (vertex *DagVertex, err error) {

	createVertexMutex.Lock()
	defer createVertexMutex.Unlock()

	err = dagStorage.queuePendingData.Push(data)
	if err != nil {
		return nil, err
	}

	createdVertex := (*DagVertex)(nil)
	if dagStorage.queuePendingData.IsFull() {

		// ComposePayloadVertex(nil)
		createdVertex, err = createVertex(dagStorage, selfNode, nil)
	}

	return createdVertex, err
}

func CreateTwoParentsVertex(dagStorage *DagStorage, selfNode *DagNode, peerParentHash []byte) (vertex *DagVertex, err error) {

	createVertexMutex.Lock()
	defer createVertexMutex.Unlock()

	peerParent := GetVertex(dagStorage, peerParentHash)
	if peerParent == nil {
		return nil, errors.New("cannot find peer parent in storage")
	}

	createdVertex, err := createVertex(dagStorage, selfNode, peerParent)

	return createdVertex, err
}

func createVertex(dagStorage *DagStorage, selfNode *DagNode, peerParent *DagVertex) (vertex *DagVertex, err error) {

	if selfNode == nil {
		log.W("[dag] createVertex: selfNode is nil.")
		return nil, errors.New("[dag] createVertex: selfNode is nil")
	}

	log.I("[dag] start to create new vertex")
	vertex = new(DagVertex)

	vertex.CreatorNodeId = selfNode.NodeId
	content := new(DagVertexContent)
	content.TimeStamp, _ = ptypes.TimestampProto(time.Now())

	// Mutex to read from storage
	lastSelfVertexHash, _ := GetLatestNodeVertex(dagStorage, selfNode.NodeId, true)
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
		return nil, errors.New("fatal: proto Marshal content failed")
	}

	sha256Hash := sha256.Sum256(contentBytes)
	vertex.Hash = sha256Hash[:]

	privateKey := selfNode.Device.PrivateKey
	signature, err := secp256k1.Sign(vertex.Hash, privateKey)
	vertex.Signature = signature

	// Save to database
	err = SaveVertex(dagStorage, vertex)

	// Clear the pending queue data
	dagStorage.queuePendingData.Clear()

	log.I("[dag] new vertex created.")
	err = nil
	return
}

func CreateGenesisVertex(dagStorage *DagStorage, selfNode *DagNode) (vertex *DagVertex, err error) {

	createVertexMutex.Lock()
	defer createVertexMutex.Unlock()

	log.I("[dag] start to create genesis vertex")
	vertex = new(DagVertex)

	vertex.CreatorNodeId = selfNode.NodeId
	content := new(DagVertexContent)
	content.TimeStamp, _ = ptypes.TimestampProto(time.Now())
	content.Salt = strconv.Itoa(rand.Intn(1000))

	// Mutex to read from storage
	content.SelfParentHash = nil
	content.PeerParentHash = nil

	vertex.Content = content

	// Calculate Hash and Encrypt it
	contentBytes, err := proto.Marshal(content)
	if err != nil {
		return nil, errors.New("fatal: proto Marshal content failed")
	}

	sha256Hash := sha256.Sum256(contentBytes)
	vertex.Hash = sha256Hash[:]

	privateKey := selfNode.Device.PrivateKey
	signature, err := secp256k1.Sign(vertex.Hash, privateKey)
	vertex.Signature = signature

	// Save to database
	err = SaveVertex(dagStorage, vertex)

	err = nil
	return
}


