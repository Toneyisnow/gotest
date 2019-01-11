package dag

import (
	"bytes"
	"crypto/sha256"
	"github.com/golang/protobuf/proto"
	"github.com/smartswarm/core/crypto/secp256k1"
	"github.com/smartswarm/go/log"
	"../storage"
)

type ProcessResult int

const (
	ProcessResult_No        = 0
	ProcessResult_Yes       = 1
	ProcessResult_Undecided = 2
)

// Choose one DagNode to send vertexes that it might not know, return nil if no need to send
func SelectPeerNodeToSendVertex(storage *DagStorage, nodes *DagNodes) (results []*DagNode) {

	return nil
}

// For a given node, find all of the unknown vertexes for it
func FindPossibleUnknownVertexesForNode(storage *DagStorage, node *DagNode) (resultList []*DagVertex, err error) {

	resultList = make([]*DagVertex, 0)

	resultList = append(resultList, nil)

	err = nil
	return
}

// ProcessIncomingVertex:
// 1. Validate the incoming vertex to be :1) signature correct, 2) hash correct. If not, return No
// 2. Make sure both of the parents are already in the Dag, otherwise return Undecided
// 3. Put the vertex into tableVertex, queueFreshVertex, levelQueueUnconfirmedVertex
func ProcessIncomingVertex(dagStorage *DagStorage, nodes *DagNodes, vertex *DagVertex) ProcessResult {

	if dagStorage == nil || vertex == nil || vertex.Hash == nil || vertex.Signature == nil ||vertex.CreatorNodeId == 0 {
		return ProcessResult_No
	}

	// Confirm the hash and signature is correct
	contentBytes, err := proto.Marshal(vertex.Content)
	if err != nil {
		log.W("Fatal: proto Marshal content failed.")
		return ProcessResult_No
	}

	calculatedHash := sha256.Sum256(contentBytes)[:]

	if !bytes.Equal(vertex.Hash, calculatedHash) {
		log.W("ProcessIncomingVertex: hash value is not correct")
		return ProcessResult_No
	}

	peerNode := nodes.GetPeerNodeById(vertex.CreatorNodeId)
	if peerNode == nil {
		log.W("ProcessIncomingVertex: cannot find creator node")
		return ProcessResult_No
	}

	calculatedSignature, err := secp256k1.Sign(calculatedHash, peerNode.Device.PrivateKey)
	if !bytes.Equal(vertex.Signature, calculatedSignature) {
		log.W("ProcessIncomingVertex: signature is not matching")
		return ProcessResult_No
	}

	// Find parents
	selfParentHash := vertex.GetContent().GetSelfParentHash()
	peerParentHash := vertex.GetContent().GetPeerParentHash()

	if selfParentHash == nil {
		log.W("ProcessIncomingVertex: self parent hash is nil")
		return ProcessResult_No
	}

	undetermined := false
	if !dagStorage.tableVertex.Exists(selfParentHash) {
		undetermined = true
	}

	if peerParentHash != nil && !dagStorage.tableVertex.Exists(peerParentHash) {
		undetermined = true
	}

	if undetermined {
		// At least one of the parent is not in the Dag, return Undecided
		return ProcessResult_Undecided
	}

	// Save the new vertex into tableVertex
	vertexBytes, err := proto.Marshal(vertex)
	if err != nil {
		log.W("ProcessIncomingVertex: marshal vertex failed.")
		return ProcessResult_No
	}
	err = dagStorage.tableVertex.InsertOrUpdate(vertex.GetHash(), vertexBytes)

	// Save to tableVertexParent
	vertexLink := &DagVertexLink{}
	vertexLink.NodeId = vertex.CreatorNodeId
	vertexLink.SelfParentHash = selfParentHash
	vertexLink.PeerParentHash = peerParentHash
	vertexParentBytes, err := proto.Marshal(vertexLink)
	err = dagStorage.tableVertexLink.InsertOrUpdate(vertex.GetHash(), vertexParentBytes)

	return ProcessResult_Yes
}

// ProcessVertexAndDecideCandidate:
// 1. Calculate the Sees() for the freshVertex vs. each of the vertexes in candidates
// 2. If the Sees() result is greater than majority of nodes for majority of candidates, then mark this vertex
// to be level+1
// 3. If this is the first vertex in the level for this node, mark this vertex as candidate and return Yes, otherwise No
// 4. Save this vertex into tableLatestVertex, tableCandidate
func ProcessVertexAndDecideCandidate(dagStorage *DagStorage, dagNodes *DagNodes, vertexHash []byte) ProcessResult {

	link := GetVertexLink(dagStorage, vertexHash)
	if link == nil {
		return ProcessResult_No
	}

	selfParentStatus := GetVertexStatus(dagStorage, link.SelfParentHash)
	peerParentStatus := GetVertexStatus(dagStorage, link.PeerParentHash)

	peerParentLink := GetVertexLink(dagStorage, link.PeerParentHash)
	if peerParentLink == nil {
		return ProcessResult_No
	}

	if selfParentStatus == nil || selfParentStatus.Level == 0 || peerParentStatus == nil || peerParentStatus.Level == 0 {

		// At least one onf the parent status is not ready yet, just return Undecided and wait for processing again
		return ProcessResult_Undecided
	}

	currentLevel := selfParentStatus.Level
	if selfParentStatus.Level < peerParentStatus.Level {
		currentLevel = peerParentStatus.Level
	}

	vertexStatus := GetVertexStatus(dagStorage, vertexHash)
	if vertexStatus == nil {
		vertexStatus = &DagVertexStatus{}
	}

	strongConnectionCount := 0
	for _, dagNode := range dagNodes.AllNodes() {

		candidateHash, _ := GetCandidateForNode(dagStorage, dagNode.NodeId, currentLevel, true)

		connection := CalculateVertexConnection(dagStorage, vertexHash, candidateHash)
		if len(connection.NodeIdList) >= dagNodes.GetMajorityCount() {
			strongConnectionCount ++
		}
	}

	if strongConnectionCount > dagNodes.GetMajorityCount() {
		vertexStatus.Level = currentLevel + 1
		vertexStatus.IsCandidate = true
	} else {
		vertexStatus.Level = currentLevel
	}

	// Save the vertex status
	vertexStatusByte, _ := proto.Marshal(vertexStatus)
	err := dagStorage.tableVertexStatus.InsertOrUpdate(vertexHash, vertexStatusByte)

	// Push to queue and channel if necessary
	err = dagStorage.levelQueueUnconfirmedVertex.Push(vertexStatus.Level, vertexHash)

	if vertexStatus.IsCandidate {
		err = dagStorage.levelQueueUndecidedCandidate.Push(vertexStatus.Level, vertexHash)
		return ProcessResult_Yes
	} else {
		return ProcessResult_No
	}
	
}

// ProcessCandidateVote:
// 1. For the new given candidate, vote for each of the undecidedCandidate whether they are queen
// 2. Save the vote results into tableVote
func ProcessCandidateVote(storage *DagStorage, freshCandidateHash []byte) ProcessResult {

	return ProcessResult_Yes
}

// Return Yes if found new queen
func ProcessCandidateCollectVote(storage *DagStorage, freshCandidateHash []byte) ProcessResult {

	return ProcessResult_Yes
}


// Queen to decide whether a vertex is Accepted/Rejected
// 1. For each vertex in unconfirmedVertexQueue, use the queen to decide
// 2. If everything goes well, return Yes
// 3. If wrong happen, return No. The upper will re-process the queen later
func ProcessQueenDecision(storage *DagStorage, queenHash []byte) ProcessResult {

	return ProcessResult_Yes
}

func GetVertexStatus(dagStorage *DagStorage, vertexHash []byte) *DagVertexStatus {

	statusByte := dagStorage.tableVertexStatus.Get(vertexHash)
	if statusByte == nil {
		return nil
	}

	status := &DagVertexStatus{}
	err := proto.Unmarshal(statusByte, status)
	if err != nil {
		return nil
	}

	return status
}

func GetVertexLink(dagStorage *DagStorage, vertexHash []byte) *DagVertexLink {

	linkByte := dagStorage.tableVertexStatus.Get(vertexHash)
	if linkByte == nil {
		return nil
	}

	link := &DagVertexLink{}
	err := proto.Unmarshal(linkByte, link)
	if err != nil {
		return nil
	}

	return link
}

func GetCandidateForNode(dagStorage *DagStorage, nodeId uint64, level uint32, hashOnly bool) (hash []byte, vertex *DagVertex) {

	key := append(storage.ConvertUint64ToBytes(nodeId), storage.ConvertUint32ToBytes(level)...)
	resultByte := dagStorage.tableVertexStatus.Get(key)
	if resultByte == nil {
		return nil, nil
	}

	if hashOnly {
		return resultByte, nil
	}

	vertexByte := dagStorage.tableVertex.Get(resultByte)

	result := &DagVertex{}
	err := proto.Unmarshal(vertexByte, result)
	if err != nil {
		return nil, nil
	}

	return resultByte, result
}

func GetVertexConnection(dagStorage *DagStorage, vertexHash []byte, targetVertexHash []byte) *DagVertexConnection {

	key := append(vertexHash, targetVertexHash...)
	resultByte := dagStorage.tableVertexStatus.Get(key)
	if resultByte == nil {
		return nil
	}

	result := &DagVertexConnection{}
	err := proto.Unmarshal(resultByte, result)
	if err != nil {
		return nil
	}

	return result
}

func CalculateVertexConnection(dagStorage *DagStorage, vertexHash []byte, targetVertexHash []byte) *DagVertexConnection {

	connectionResult := &DagVertexConnection{}


	return connectionResult
}