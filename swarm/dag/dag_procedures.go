package dag

import (
	"../storage"
	"bytes"
	"encoding/hex"
	"github.com/golang/protobuf/proto"
	"github.com/smartswarm/go/log"
	"math/rand"
	"time"
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

type VertexConfirmResult int
const (
	VertexConfirmResult_Accepted = 1
	VertexConfirmResult_Rejected = 2
)

// Choose one DagNode to send vertexes that it might not know, return nil if no need to send
func SelectPeerNodeToSendVertex(dagStorage *DagStorage, vertex *DagVertex, dagNodes *DagNodes) (results []*DagNode) {

	log.I("[dag] selecting peer nodes to send vertex...")

	if dagStorage == nil || vertex == nil || dagNodes == nil {

		return nil
	}

	results = make([]*DagNode, 0)
	nodeNeeded := 0
	if vertex.GetContent().PeerParentHash != nil {
		// If this vertex is created from peer triggering, just send to 0-1 other nodes
		if rand.Intn(100) < 50 {
			nodeNeeded = 1
		}
	} else {
		// If this vertex is from itself, send to 2 other nodes
		nodeNeeded = 2
	}
	log.I("[dag] node needed=", nodeNeeded)

	for _, peerNode := range dagNodes.Peers {
		if nodeNeeded == 0 {
			break
		}

		key := storage.ConvertUint64ToBytes(peerNode.NodeId)
		lastTime := time.Time{}
		lastTimeBytes := dagStorage.tableNodeSyncTimestamp.Get(key)

		if lastTimeBytes != nil {
			err := lastTime.UnmarshalBinary(lastTimeBytes)
			if err != nil {
				// do nothing here
			}
		}

		// The cooldown time for a given peer node to sync with current node is 1 second
		if time.Now().Sub(lastTime) > time.Second {

			results = append(results, peerNode)
			nodeNeeded --

			b, _ := time.Now().MarshalBinary()
			err := dagStorage.tableNodeSyncTimestamp.InsertOrUpdate(key, b)
			if err != nil {
				// do nothing here
			}
		}
	}

	return results
}

// For a given node, find all of the unknown vertexes for it
func FindPossibleUnknownVertexesForNode(dagStorage *DagStorage, selfNode *DagNode,  peerNode *DagNode) (resultList []*DagVertex, err error) {

	resultList = make([]*DagVertex, 0)

	if dagStorage == nil || selfNode == nil || peerNode == nil {
		return
	}

	_, latestVertex := GetNodeLatestVertex(dagStorage, selfNode.NodeId, false)
	if latestVertex == nil {
		return
	}
	log.I("[dag] latest vertex on node ", selfNode.NodeId, ": vertex=", GetShortenedHash(latestVertex.Hash))

	potentialDependentList := []*DagVertex { latestVertex }

	for {
		if len(potentialDependentList) == 0 {
			log.I("[dag] potential dependent list is empty, break it.")
			break
		}

		newDependents := make([]*DagVertex, 0)

		for _, dVertex := range potentialDependentList {

			log.I("[dag] iterating vertex", hex.EncodeToString(dVertex.Hash))
			if dVertex == nil || dVertex.GetContent() == nil {
				log.W("[dag] potential vertex is broken. ignore it.")
				continue
			}

			if !DoesExistNodeSyncVertex(dagStorage, peerNode.NodeId, dVertex.Hash) {

				log.I("[dag] node", peerNode.NodeId, "does not know vertex", GetShortenedHash(dVertex.Hash), ", adding it to related list.")
				resultList = append([]*DagVertex{ dVertex }, resultList...)

				if dVertex.GetContent().SelfParentHash != nil {

					selfParentVertex := GetVertex(dagStorage, dVertex.GetContent().SelfParentHash)
					newDependents = append(newDependents, selfParentVertex)
				}
				if dVertex.GetContent().PeerParentHash != nil {

					selfParentVertex := GetVertex(dagStorage, dVertex.GetContent().PeerParentHash)
					newDependents = append(newDependents, selfParentVertex)
				}
			}
		}

		potentialDependentList = newDependents
	}

	err = nil
	return
}

func FlagKnownVertexForNode(dagStorage *DagStorage, node *DagNode, vertexList []*DagVertex) {

	if dagStorage == nil || node == nil || vertexList == nil {
		return
	}

	for _, vertex := range vertexList {
		SetNodeSyncVertex(dagStorage, node.NodeId, vertex.Hash)
	}
}

// ProcessIncomingVertex:
// 1. Make sure both of the parents are already in the Dag, otherwise return Undecided
// 2. Build the vertex into graph
//
func ProcessIncomingVertex(dagStorage *DagStorage, nodes *DagNodes, vertexHash []byte) (result ProcessResult, missingParentHash []byte) {

	if dagStorage == nil || vertexHash == nil {
		return ProcessResult_No, nil
	}

	log.I("[dag] process incoming vertex:", GetShortenedHash(vertexHash))
	vertex := GetVertex(dagStorage, vertexHash)

	if vertex == nil || vertex.Hash == nil || vertex.Signature == nil ||vertex.CreatorNodeId == 0 || vertex.GetContent() == nil {
		log.W("[dag] process incoming vertex: vertex is broken.")
		return ProcessResult_No, nil
	}

	creatorNode := nodes.GetPeerNodeById(vertex.CreatorNodeId)
	if creatorNode == nil {
		log.W("[dag] process incoming vertex: cannot find creator node")
		return ProcessResult_No, nil
	}

	// Find parents
	selfParentHash := vertex.GetContent().GetSelfParentHash()
	peerParentHash := vertex.GetContent().GetPeerParentHash()

	if selfParentHash == nil {

		log.I("[dag] self parent hash is nil")
		// Check if this is genesisVertex
		genesisVertexHash, _ := GetGenesisVertex(dagStorage, creatorNode.NodeId, true)
		if genesisVertexHash == nil || bytes.Equal(genesisVertexHash, vertex.Hash) {

			if genesisVertexHash == nil {
				log.I("[dag] cannot find genesis vertex for node [", creatorNode.NodeId, "], assigning this vertex as genesis.")
				// This is the genesis vertex, save it
				SetGenesisVertex(dagStorage, creatorNode.NodeId, vertex.Hash)
			}

			// Building the genesis vertex
			if BuildVertexGraph(dagStorage, creatorNode.NodeId, vertex.Hash, nil, nil) {

				return ProcessResult_Yes, nil
			} else {
				// Something temporary failed, should retry
				return ProcessResult_Undecided, nil
			}

		} else {

			// This is not genesis vertex, should error out
			log.W("ProcessIncomingVertex: self parent hash is nil and it's not genesis vertex")
			return ProcessResult_No, nil
		}
	}

	if GetVertex(dagStorage, selfParentHash) == nil || GetVertexLink(dagStorage, selfParentHash) == nil {
		return ProcessResult_Undecided, selfParentHash
	}

	if peerParentHash != nil &&
		(GetVertex(dagStorage, peerParentHash) == nil || GetVertexLink(dagStorage, peerParentHash) == nil) {
		return ProcessResult_Undecided, peerParentHash
	}

	//Save the vertex to graph
	if BuildVertexGraph(dagStorage, vertex.CreatorNodeId, vertex.Hash, selfParentHash, peerParentHash) {
		return ProcessResult_Yes, nil
	} else {
		return ProcessResult_Undecided, nil
	}
}

// ProcessVertexAndDecideCandidate:
// 1. Calculate the Sees() for the freshVertex vs. each of the vertexes in candidates
// 2. If the Sees() result is greater than majority of nodes for majority of candidates, then mark this vertex
// to be level+1
// 3. If this is the first vertex in the level for this node, mark this vertex as candidate and return Yes, otherwise No
// 4. Save this vertex into tableLatestVertex, tableCandidate
//
// Table to fill: tableVertexStatus,
func ProcessVertexAndDecideCandidate(dagStorage *DagStorage, dagNodes *DagNodes, vertexHash []byte) (result ProcessResult, missingParentHash []byte) {

	log.I("[dag] processing vertex and decide candidate... vertex=", GetShortenedHash(vertexHash))

	link := GetVertexLink(dagStorage, vertexHash)
	if link == nil {
		log.W("[dag] cannot find vertex link for vertex.")
		return ProcessResult_No, nil
	}

	if link.SelfParentHash == nil {

		// This should be a genesis vertex, double check for each nodes to confirm.
		isGenesis, nodeId := IsGenesisVertex(dagStorage, dagNodes, vertexHash)
		if isGenesis {

			log.I("[dag] the vertex is genesis vertex, push it into candidate queue.")

			// Set the node Level=1, isCandidate=true
			vertexStatus := &DagVertexStatus{ Level:1, IsCandidate:true, IsQueen:false, IsQueenDecided:false }
			SetVertexStatus(dagStorage, vertexHash, vertexStatus)

			SetCandidateForNode(dagStorage, nodeId, uint32(1), vertexHash)

			err := dagStorage.levelQueueUnconfirmedVertex.Push(uint32(1), vertexHash)
			err = dagStorage.levelQueueUndecidedCandidate.Push(uint32(1), vertexHash)
			if err != nil {

			}
			return ProcessResult_Yes, nil
		} else {
			log.W("[dag] self parent hash is nil and vertex is not genesis, stop processing it.")
			return ProcessResult_No, nil
		}
	}

	// If self parent status is not ready yet, just return Undecided and wait for processing again
	selfParentStatus := GetVertexStatus(dagStorage, link.SelfParentHash)
	if selfParentStatus == nil || selfParentStatus.Level == 0 {
		log.W("[dag] cannot find self parent status, postpone it to next execution.")
		return ProcessResult_Undecided, link.SelfParentHash
	}

	currentLevel := selfParentStatus.Level

	if link.PeerParentHash != nil {
		peerParentStatus := GetVertexStatus(dagStorage, link.PeerParentHash)
		if peerParentStatus == nil || peerParentStatus.Level == 0 {
			log.W("[dag] cannot find peer parent status, postpone it to next execution.")
			return ProcessResult_Undecided, link.PeerParentHash
		}

		peerParentLink := GetVertexLink(dagStorage, link.PeerParentHash)
		if peerParentLink == nil {
			// This should not happen
			log.W("[dag] error while getting peer parent link.")
			return ProcessResult_No, nil
		}

		if selfParentStatus.Level < peerParentStatus.Level {
			currentLevel = peerParentStatus.Level
		}
	}

	log.I("[dag] current level of the vertex:", currentLevel)

	vertexStatus := GetVertexStatus(dagStorage, vertexHash)
	if vertexStatus == nil {
		vertexStatus = &DagVertexStatus{}
	}

	strongConnectionCount := 0
	for _, dagNode := range dagNodes.AllNodes() {

		candidateHash, _ := GetCandidateForNode(dagStorage, dagNode.NodeId, currentLevel, true)
		if candidateHash == nil {
			// It's possible that for given node and level there is no candidate
			continue
		}

		log.I("[dag] got candidate on node=", dagNode.NodeId, "level=", currentLevel, "vertex_hash=", candidateHash)
		connection := CalculateVertexConnection(dagStorage, vertexHash, candidateHash)
		if connection == nil {
			// Data is missing, so skip this connection for now
			continue
		}

		if connection.IsConnected() && connection.GetNodeCount() >= dagNodes.GetMajorityCount() {
			log.I("[dag] connection from", GetShortenedHash(vertexHash), "to candidate", GetShortenedHash(candidateHash), "is strong.")
			strongConnectionCount ++
		} else {
			log.I("[dag] connection from", GetShortenedHash(vertexHash), "to candidate", GetShortenedHash(candidateHash), "is not strong.")
		}
	}
	log.I("[dag] strong connection count:", strongConnectionCount)

	if strongConnectionCount > dagNodes.GetMajorityCount() {
		vertexStatus.Level = currentLevel + 1
		vertexStatus.IsCandidate = true
	} else {
		vertexStatus.Level = currentLevel
	}

	log.I("[dag] set status for vertex. level=", vertexStatus.Level, "isCandidate=",vertexStatus.IsCandidate, "vertex=", GetShortenedHash(vertexHash))

	// Save the vertex status
	SetVertexStatus(dagStorage, vertexHash, vertexStatus)

	// Push to queue and channel if necessary
	err := dagStorage.levelQueueUnconfirmedVertex.Push(vertexStatus.Level, vertexHash)
	if err != nil {
		// temporary fail, should retry
		return ProcessResult_Undecided, nil
	}

	if vertexStatus.IsCandidate {

		SetCandidateForNode(dagStorage, link.NodeId, vertexStatus.Level, vertexHash)
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

	log.I("[dag] processing candidate vote. candidate hash=", GetShortenedHash(nowCandidateHash))
	nowCandidateStatus := GetVertexStatus(dagStorage, nowCandidateHash)

	if nowCandidateStatus == nil {
		return ProcessResult_No
	}

	newQueenUpdated := ProcessResult(ProcessResult_No)

	dagStorage.levelQueueUndecidedCandidate.StartIterate()
	targetCandidateIndex, targetCandidateHash := dagStorage.levelQueueUndecidedCandidate.IterateNext()
	for targetCandidateHash != nil {

		targetCandidateStatus := GetVertexStatus(dagStorage, targetCandidateHash)
		if nowCandidateStatus.Level >= targetCandidateStatus.Level + 10 {

			// Coin round to decide Yes or No
			// TODO: just put No Decision for now
			SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_DecideNo)
			dagStorage.levelQueueUndecidedCandidate.Delete(targetCandidateIndex)

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
				targetCandidateStatus.IsQueenDecided = true
				targetCandidateStatus.IsQueen = true
				SetVertexStatus(dagStorage, targetCandidateHash, targetCandidateStatus)

				onQueenFound(targetCandidateHash)
				dagStorage.levelQueueUndecidedCandidate.Delete(targetCandidateIndex)
				newQueenUpdated = ProcessResult_Yes

			} else if noCount >= dagNodes.GetMajorityCount() {
				SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_DecideNo)

				// Change the candidate to queen
				targetCandidateStatus.IsQueenDecided = true
				targetCandidateStatus.IsQueen = false
				SetVertexStatus(dagStorage, targetCandidateHash, targetCandidateStatus)

				onQueenFound(targetCandidateHash)
				dagStorage.levelQueueUndecidedCandidate.Delete(targetCandidateIndex)
				newQueenUpdated = ProcessResult_Yes

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

		targetCandidateIndex, targetCandidateHash = dagStorage.levelQueueUndecidedCandidate.IterateNext();
	}

	return newQueenUpdated
}

// Queen to decide whether a vertex is Accepted/Rejected
// 1. For each vertex in unconfirmedVertexQueue, use the queen to decide
// 2. If everything goes well, return Yes
// 3. If wrong happen, return No. The upper will re-process the queen later
func ProcessQueenDecision(dagStorage *DagStorage, dagNodes *DagNodes, queenHash []byte,
	vertexConfirmer func (vertexHash []byte, result VertexConfirmResult)) ProcessResult {

	queenStatus := GetVertexStatus(dagStorage, queenHash)
	if queenStatus == nil {
		return ProcessResult_No
	}

	allCandidatesDecided := true
	allQueensInLevel := make([][]byte, 0)
	for _, node := range dagNodes.AllNodes() {

		candidateHash, _ := GetCandidateForNode(dagStorage, node.NodeId, queenStatus.Level, true)
		if candidateHash == nil {
			continue
		}

		candidateStatus := GetVertexStatus(dagStorage, candidateHash)
		if candidateStatus == nil || !candidateStatus.IsQueenDecided {
			allCandidatesDecided = false
			break
		}

		if candidateStatus.IsQueen {
			allQueensInLevel = append(allQueensInLevel, candidateHash)
		}
	}

	if !allCandidatesDecided || len(allQueensInLevel) == 0 {

		// Not all candidates on this level decided is queen or not, or there is no queen in this level,
		// just pass this and do nothing
		return ProcessResult_Yes
	}

	// Start iterating all the vertexes, and using the AllQueensInLevel to decide accept/reject it
	dagStorage.levelQueueUnconfirmedVertex.StartIterate()
	targetVertexIndex, targetVertexHash := dagStorage.levelQueueUnconfirmedVertex.IterateNext()
	for targetVertexHash != nil {

		hasAllConnection := true
		for _, iQueenHash := range allQueensInLevel {

			connection := GetVertexConnection(dagStorage, iQueenHash, targetVertexHash)
			if connection == nil || len(connection.NodeIdList) == 0 {
				hasAllConnection = false
				break
			}
		}

		if hasAllConnection {

			vertexConfirmer(targetVertexHash, VertexConfirmResult_Accepted)
			dagStorage.levelQueueUnconfirmedVertex.Delete(targetVertexIndex)
		}

		targetVertexIndex, targetVertexHash = dagStorage.levelQueueUnconfirmedVertex.IterateNext()
	}

	return ProcessResult_Yes
}

func EnsureGenesisVertex(dagStorage *DagStorage, node *DagNode) (newCreated bool, hash []byte) {

	genesisVertexHash, _ := GetGenesisVertex(dagStorage, node.NodeId, true)

	if genesisVertexHash != nil {
		return false, genesisVertexHash
	}

	// Create if missing
	vertex, err := CreateGenesisVertex(dagStorage, node)
	if err != nil {
		return false, nil
	}

	SetGenesisVertex(dagStorage, node.NodeId, vertex.Hash)
	SetNodeLatestVertex(dagStorage, node.NodeId, vertex.Hash)

	return true, vertex.Hash
}

func BuildVertexGraph(dagStorage *DagStorage, nodeId uint64, vertexHash []byte, selfParentHash []byte, peerParentHash []byte) bool {

	if dagStorage == nil || vertexHash == nil || nodeId == 0 {
		log.W("[dag] cannot build vertex graph.")
		return false
	}

	// Save to tableVertexLink
	vertexLink := &DagVertexLink{}
	vertexLink.NodeId = nodeId
	vertexLink.SelfParentHash = selfParentHash
	vertexLink.PeerParentHash = peerParentHash
	vertexLinkBytes, _ := proto.Marshal(vertexLink)
	err := dagStorage.tableVertexLink.InsertOrUpdate(vertexHash, vertexLinkBytes)
	if err != nil {
		return false
	}

	// Update the Latest Vertex on Node
	SetNodeLatestVertex(dagStorage, nodeId, vertexHash)

	return true
}

//
// CalculateVertexConnection:
// This should always return a non-null value:
//
func CalculateVertexConnection(dagStorage *DagStorage, vertexHash []byte, targetVertexHash []byte) *DagVertexConnection {

	vertexLink := GetVertexLink(dagStorage, vertexHash)
	if vertexLink == nil {
		// Something wrong unexpected, just return nil
		return nil
	}

	existingConnection := GetOrDefaultVertexConnection(dagStorage, vertexHash, targetVertexHash, vertexLink.NodeId)
	if existingConnection != nil {
		return existingConnection
	}

	nodeIdList := make([]uint64, 0)

	// Consider the genesis vertex which self parent is nil, so adding this check
	if vertexLink.SelfParentHash != nil {
		selfParentConnection := GetOrDefaultVertexConnection(dagStorage, vertexLink.SelfParentHash, targetVertexHash, vertexLink.NodeId)
		if selfParentConnection == nil {
			// Self parent connection is not ready yet, wait for next round to process
			return nil
		} else {
			nodeIdList = selfParentConnection.NodeIdList
		}
	}

	if vertexLink.PeerParentHash != nil {

		peerParentLink := GetVertexLink(dagStorage, vertexLink.PeerParentHash)
		if peerParentLink != nil {
			peerParentConnection := GetOrDefaultVertexConnection(dagStorage, vertexLink.PeerParentHash, targetVertexHash, peerParentLink.NodeId)

			if peerParentConnection == nil {
				// Peer parent connection is not ready yet, wait for next round to process
				return nil
			} else {
				nodeIdList = MergeUint64Array(nodeIdList, peerParentConnection.NodeIdList)
			}
		}
	}

	connectionResult := NewDagVertexConnection()

	// merge with current node only if the previous merged list is not empty
	if len(nodeIdList) > 0 {
		connectionResult.NodeIdList = MergeUint64Array(nodeIdList, []uint64{ vertexLink.NodeId })
	}

	// Save it into storage
	SetVertexConnection(dagStorage, vertexHash, targetVertexHash, connectionResult)

	return connectionResult
}

func GetOrDefaultVertexConnection(dagStorage *DagStorage, vertexHash []byte, targetVertexHash []byte, nodeId uint64) *DagVertexConnection {

	if bytes.Equal(vertexHash, targetVertexHash) {
		connectionResult := &DagVertexConnection{}
		connectionResult.NodeIdList = []uint64 { nodeId }
		return connectionResult

	} else {
		return GetVertexConnection(dagStorage, vertexHash, targetVertexHash)
	}
}

func IsGenesisVertex(dagStorage *DagStorage, dagNodes *DagNodes, vertexHash []byte) (bool, uint64) {

	for _, node := range dagNodes.AllNodes() {

		genesisVertex, _ := GetGenesisVertex(dagStorage, node.NodeId, true)

		if genesisVertex == nil {
			continue
		}

		if bytes.Equal(genesisVertex, vertexHash) {
			return true, node.NodeId
		}
	}

	return false, 0
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

func GetShortenedHash(hash []byte) string {

	hashString := hex.EncodeToString(hash)

	return hashString[:5] + "..." + hashString[len(hashString)-2:]
}