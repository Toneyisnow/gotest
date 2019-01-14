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

type CandidateDecision int
const (
	CandidateDecision_Unknown   = 0
	CandidateDecision_No	    = 1
	CandidateDecision_Yes       = 2
	CandidateDecision_DecideNo  = 3
	CandidateDecision_DecideYes = 4
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
func ProcessIncomingVertex(dagStorage *DagStorage, nodes *DagNodes, vertex *DagVertex) (result ProcessResult, missingParentHash []byte) {

	if dagStorage == nil || vertex == nil || vertex.Hash == nil || vertex.Signature == nil ||vertex.CreatorNodeId == 0 {
		return ProcessResult_No, nil
	}

	// Confirm the hash and signature is correct
	contentBytes, err := proto.Marshal(vertex.Content)
	if err != nil {
		log.W("Fatal: proto Marshal content failed.")
		return ProcessResult_No, nil
	}

	calculatedHash := sha256.Sum256(contentBytes)[:]

	if !bytes.Equal(vertex.Hash, calculatedHash) {
		log.W("ProcessIncomingVertex: hash value is not correct")
		return ProcessResult_No, nil
	}

	peerNode := nodes.GetPeerNodeById(vertex.CreatorNodeId)
	if peerNode == nil {
		log.W("ProcessIncomingVertex: cannot find creator node")
		return ProcessResult_No, nil
	}

	calculatedSignature, err := secp256k1.Sign(calculatedHash, peerNode.Device.PrivateKey)
	if !bytes.Equal(vertex.Signature, calculatedSignature) {
		log.W("ProcessIncomingVertex: signature is not matching")
		return ProcessResult_No, nil
	}

	// Find parents
	selfParentHash := vertex.GetContent().GetSelfParentHash()
	peerParentHash := vertex.GetContent().GetPeerParentHash()

	if selfParentHash == nil {
		log.W("ProcessIncomingVertex: self parent hash is nil")
		return ProcessResult_No, nil
	}

	if !dagStorage.tableVertex.Exists(selfParentHash) {
		return ProcessResult_Undecided, selfParentHash
	}

	if peerParentHash != nil && !dagStorage.tableVertex.Exists(peerParentHash) {
		return ProcessResult_Undecided, peerParentHash
	}

	// Save the new vertex into tableVertex
	vertexBytes, err := proto.Marshal(vertex)
	if err != nil {
		log.W("ProcessIncomingVertex: marshal vertex failed.")
		return ProcessResult_No, nil
	}
	err = dagStorage.tableVertex.InsertOrUpdate(vertex.GetHash(), vertexBytes)

	// Save to tableVertexParent
	vertexLink := &DagVertexLink{}
	vertexLink.NodeId = vertex.CreatorNodeId
	vertexLink.SelfParentHash = selfParentHash
	vertexLink.PeerParentHash = peerParentHash
	vertexParentBytes, err := proto.Marshal(vertexLink)
	err = dagStorage.tableVertexLink.InsertOrUpdate(vertex.GetHash(), vertexParentBytes)

	return ProcessResult_Yes, nil
}

// ProcessVertexAndDecideCandidate:
// 1. Calculate the Sees() for the freshVertex vs. each of the vertexes in candidates
// 2. If the Sees() result is greater than majority of nodes for majority of candidates, then mark this vertex
// to be level+1
// 3. If this is the first vertex in the level for this node, mark this vertex as candidate and return Yes, otherwise No
// 4. Save this vertex into tableLatestVertex, tableCandidate
func ProcessVertexAndDecideCandidate(dagStorage *DagStorage, dagNodes *DagNodes, vertexHash []byte) (result ProcessResult, missingParentHash []byte) {

	link := GetVertexLink(dagStorage, vertexHash)
	if link == nil {
		return ProcessResult_No, nil
	}

	selfParentStatus := GetVertexStatus(dagStorage, link.SelfParentHash)
	peerParentStatus := GetVertexStatus(dagStorage, link.PeerParentHash)

	peerParentLink := GetVertexLink(dagStorage, link.PeerParentHash)
	if peerParentLink == nil {
		return ProcessResult_No, nil
	}

	if selfParentStatus == nil || selfParentStatus.Level == 0 {
		// At least one onf the parent status is not ready yet, just return Undecided and wait for processing again
		return ProcessResult_Undecided, link.SelfParentHash
	}

	if peerParentStatus == nil || peerParentStatus.Level == 0 {

		// At least one onf the parent status is not ready yet, just return Undecided and wait for processing again
		return ProcessResult_Undecided, link.PeerParentHash
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
	if err != nil {
		return ProcessResult_No, nil
	}

	// Push to queue and channel if necessary
	err = dagStorage.levelQueueUnconfirmedVertex.Push(vertexStatus.Level, vertexHash)

	if vertexStatus.IsCandidate {
		err = dagStorage.levelQueueUndecidedCandidate.Push(vertexStatus.Level, vertexHash)
		return ProcessResult_Yes, nil
	} else {
		return ProcessResult_No, nil
	}
}

// ProcessCandidateVote:
// 1. For the new given candidate, vote for each of the undecidedCandidate with level-1 whether they are queen
// 2. Iterate all undecidedCandidate, collect the votes from level-1
func ProcessCandidateVote(dagStorage *DagStorage, dagNodes *DagNodes, nowCandidateHash []byte, onQueenFound func([]byte)) ProcessResult {

	nowCandidateStatus := GetVertexStatus(dagStorage, nowCandidateHash)

	if nowCandidateStatus == nil {
		return ProcessResult_No
	}

	newQueenFound := ProcessResult(ProcessResult_No)

	dagStorage.levelQueueUndecidedCandidate.StartIterate()
	targetCandidateHash := dagStorage.levelQueueUndecidedCandidate.IterateNext()
	for targetCandidateHash != nil {

		targetCandidateStatus := GetVertexStatus(dagStorage, targetCandidateHash)
		if nowCandidateStatus.Level >= targetCandidateStatus.Level + 10 {

			// Coin round to decide Yes or No
			// TODO: just put No Decision for now
			SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_DecideNo)
			dagStorage.levelQueueUndecidedCandidate.Pop()

		} else if nowCandidateStatus.Level > targetCandidateStatus.Level + 1 {

			// Collect the vote results
			yesCount := 0
			noCount := 0
			subLevel := nowCandidateStatus.Level - 1
			for _, node := range dagNodes.AllNodes() {
				subCandidateHash, _ := GetCandidateForNode(dagStorage, node.NodeId, subLevel, true)

				// Only collect decisions from strong connected candidates
				connection := GetVertexConnection(dagStorage, nowCandidateHash, subCandidateHash)
				if len(connection.NodeIdList) < dagNodes.GetMajorityCount() {
					continue
				}

				subDecision := GetCandidateDecision(dagStorage, subCandidateHash, targetCandidateHash)
				switch subDecision {
				case CandidateDecision_No:
				case CandidateDecision_DecideNo:
				case CandidateDecision_Unknown:
					noCount ++
					break;
				case CandidateDecision_Yes:
					yesCount ++
					break;
				case CandidateDecision_DecideYes:
					// This will not happen, since it's already decided by the Decision Yes
					break;
				}
			}

			if yesCount >= dagNodes.GetMajorityCount() {
				SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_DecideYes)

				// Change the candidate to queen
				targetCandidateStatus.IsQueen = true
				SetVertexStatus(dagStorage, targetCandidateHash, targetCandidateStatus)

				onQueenFound(targetCandidateHash)
				dagStorage.levelQueueUndecidedCandidate.Pop()
				newQueenFound = ProcessResult_Yes

			} else if noCount >= dagNodes.GetMajorityCount() {
				SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_DecideNo)
			} else if yesCount > noCount {
				SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_Yes)
			} else if yesCount < noCount {
				SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_No)
			} else {
				SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_Unknown)
			}

		} else if nowCandidateStatus.Level == targetCandidateStatus.Level + 1 {

			// Vote the candidate
			connection := GetVertexConnection(dagStorage, nowCandidateHash, targetCandidateHash)
			if connection != nil && len(connection.NodeIdList) > 0 {
				SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_Yes)
			}
		}

		targetCandidateHash = dagStorage.levelQueueUndecidedCandidate.IterateNext();
	}

	return newQueenFound
}

// Queen to decide whether a vertex is Accepted/Rejected
// 1. For each vertex in unconfirmedVertexQueue, use the queen to decide
// 2. If everything goes well, return Yes
// 3. If wrong happen, return No. The upper will re-process the queen later
func ProcessQueenDecision(dagStorage *DagStorage, queenHash []byte) ProcessResult {

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

func SetVertexStatus(dagStorage *DagStorage, vertexHash []byte, status *DagVertexStatus) {

	statusByte, _ := proto.Marshal(status)
	dagStorage.tableVertexStatus.InsertOrUpdate(vertexHash, statusByte)
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

	connectionResult := GetVertexConnection(dagStorage, vertexHash, targetVertexHash)
	if connectionResult != nil {
		return connectionResult
	}

	connectionResult = &DagVertexConnection{}

	vertexLink := GetVertexLink(dagStorage, vertexHash)
	if vertexLink == nil || vertexLink.SelfParentHash == nil || vertexLink.PeerParentHash == nil {
		// Something wrong unexpected, just return nil
		return nil
	}

	selfParentConnection := GetVertexConnection(dagStorage, vertexLink.SelfParentHash, targetVertexHash)
	peerParentConnection := GetVertexConnection(dagStorage, vertexLink.PeerParentHash, targetVertexHash)

	connectionResult.NodeIdList = MergeUint64Array(selfParentConnection.NodeIdList, peerParentConnection.NodeIdList)
	connectionResult.NodeIdList = MergeUint64Array(connectionResult.NodeIdList, []uint64{ vertexLink.NodeId })

	return connectionResult
}

func GetCandidateDecision(dagStorage *DagStorage, vertexHash []byte, targetVertexHash []byte) CandidateDecision {

	key := append(vertexHash, targetVertexHash...)
	resultByte := dagStorage.tableCandidateDecision.Get(key)
	if resultByte == nil {
		return CandidateDecision_Unknown
	}

	result := storage.ConvertBytesToUint32(resultByte)
	return CandidateDecision(result)
}


func SetCandidateDecision(dagStorage *DagStorage, vertexHash []byte, targetVertexHash []byte, decision CandidateDecision) {

	key := append(vertexHash, targetVertexHash...)
	value := storage.ConvertUint32ToBytes(uint32(decision))

	dagStorage.tableCandidateDecision.InsertOrUpdate(key, value)
}


func MergeUint64Array(array1 []uint64, array2 []uint64) []uint64 {

	if array1 == nil {
		return array2
	}

	if array2 == nil {
		return array1
	}

	result := make([]uint64, 0)

	index1 := 0
	index2 := 0

	for index1 < len(array1) || index2 < len(array2) {

		if index1 >= len(array1) || array1[index1] > array2[index2] {
			result = append(result, array2[index2])
			index2 ++
		} else if index2 >= len(array2) || array2[index2] > array1[index1] {
			result = append(result, array1[index1])
			index1 ++
		} else {
			// They are equal, just pick one value
			result = append(result, array1[index1])
			index1 ++
			index2 ++
		}
	}

	return result
}
