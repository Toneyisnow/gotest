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

	log.D("[dag][select peer node] start...")

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
	log.D("[dag][select peer node] node needed=", nodeNeeded)

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

	log.D("[dag][find possible unknown vertex for node] start...")

	resultList = make([]*DagVertex, 0)

	if dagStorage == nil || selfNode == nil || peerNode == nil {
		return
	}

	_, latestVertex := GetNodeLatestVertex(dagStorage, selfNode.NodeId, false)
	if latestVertex == nil {
		return
	}
	log.D("[dag][find possible unknown vertex for node] latest vertex on node ", selfNode.NodeId, ": vertex=", GetShortenedHash(latestVertex.Hash))

	potentialDependentList := []*DagVertex { latestVertex }

	for {
		if len(potentialDependentList) == 0 {
			//log.I("[dag] potential dependent list is empty, break it.")
			break
		}

		newDependents := make([]*DagVertex, 0)

		for _, dVertex := range potentialDependentList {

			// log.I("[dag] iterating vertex", hex.EncodeToString(dVertex.Hash))
			if dVertex == nil || dVertex.GetContent() == nil {
				log.W("[dag][find possible unknown vertex for node] potential vertex is broken. ignore it.")
				continue
			}

			if !DoesExistNodeSyncVertex(dagStorage, peerNode.NodeId, dVertex.Hash) {

				log.D("[dag][find possible unknown vertex for node] node", peerNode.NodeId, "does not know vertex", GetShortenedHash(dVertex.Hash), ", adding it to related list.")
				resultList = append([]*DagVertex{ dVertex }, resultList...)

				if dVertex.GetContent().SelfParentHash != nil {

					selfParentVertex := GetVertex(dagStorage, dVertex.GetContent().SelfParentHash)
					newDependents = MergeVertexes(newDependents, selfParentVertex)
				}
				if dVertex.GetContent().PeerParentHash != nil {

					peerParentVertex := GetVertex(dagStorage, dVertex.GetContent().PeerParentHash)
					newDependents = MergeVertexes(newDependents, peerParentVertex)
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

	log.I("[dag][process incoming vertex] started. vertex=", GetShortenedHash(vertexHash))
	vertex := GetVertex(dagStorage, vertexHash)

	if vertex == nil || vertex.Hash == nil || vertex.Signature == nil ||vertex.CreatorNodeId == 0 || vertex.GetContent() == nil {
		log.W("[dag][process incoming vertex] vertex is broken.")
		return ProcessResult_No, nil
	}

	creatorNode := nodes.GetNodeById(vertex.CreatorNodeId)
	if creatorNode == nil {
		log.W("[dag][process incoming vertex] cannot find creator node")
		return ProcessResult_No, nil
	}

	// Find parents
	selfParentHash := vertex.GetContent().GetSelfParentHash()
	peerParentHash := vertex.GetContent().GetPeerParentHash()
	log.I("[dag][process incoming vetex] creator node=", creatorNode.NodeId, " self parent=", GetShortenedHash(selfParentHash), "peer parent=", GetShortenedHash(peerParentHash))

	if selfParentHash == nil {

		// Check if this is genesisVertex
		genesisVertexHash, _ := GetGenesisVertex(dagStorage, creatorNode.NodeId, true)
		if genesisVertexHash == nil || bytes.Equal(genesisVertexHash, vertex.Hash) {

			if genesisVertexHash == nil {
				log.D("[dag][process incoming vertex] cannot find genesis vertex for node [", creatorNode.NodeId, "], assigning this vertex as genesis.")
				// This is the genesis vertex, save it
				SetGenesisVertex(dagStorage, creatorNode.NodeId, vertex.Hash)
			}

			// Building the genesis vertex
			if BuildVertexGraph(dagStorage, creatorNode.NodeId, vertex.Hash, nil, nil) {
				log.D("[dag][process incoming vertex] successfully build genesis vertex into graph")
				return ProcessResult_Yes, nil
			} else {
				// Something temporary failed, should retry
				log.D("[dag][process incoming vertex] failed build genesis vertex, will try later")
				return ProcessResult_Undecided, nil
			}

		} else {

			// This is not genesis vertex, should error out
			log.W("[dag][process incoming vertex] self parent hash is nil and it's not genesis vertex")
			return ProcessResult_No, nil
		}
	}

	if GetVertex(dagStorage, selfParentHash) == nil || GetVertexLink(dagStorage, selfParentHash) == nil {
		log.D("[dag][process incoming vertex] self parent or its link is not ready. self parent hash=", GetShortenedHash(selfParentHash))
		return ProcessResult_Undecided, selfParentHash
	}

	if peerParentHash != nil &&
		(GetVertex(dagStorage, peerParentHash) == nil || GetVertexLink(dagStorage, peerParentHash) == nil) {
		log.D("[dag][process incoming vertex] peer parent or its link is not ready. peer parent hash=", GetShortenedHash(peerParentHash))
		return ProcessResult_Undecided, peerParentHash
	}

	//Save the vertex to graph
	if BuildVertexGraph(dagStorage, vertex.CreatorNodeId, vertex.Hash, selfParentHash, peerParentHash) {
		log.D("[dag][process incoming vertex] successfully build vertex into graph")
		return ProcessResult_Yes, nil
	} else {
		log.D("[dag][process incoming vertex] failed build vertex into graph, will try later")
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

	log.I("[dag][process vertex decide candidate] started. vertex=", GetShortenedHash(vertexHash))

	link := GetVertexLink(dagStorage, vertexHash)
	if link == nil {
		log.W("[dag][process vertex decide candidate] cannot find vertex link for vertex.")
		return ProcessResult_No, nil
	}

	if link.SelfParentHash == nil {

		// This should be a genesis vertex, double check for each nodes to confirm.
		isGenesis, nodeId := IsGenesisVertex(dagStorage, dagNodes, vertexHash)
		if isGenesis {

			log.D("[dag][process vertex decide candidate] the vertex is genesis vertex, push it into candidate queue.")

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
			log.W("[dag][process vertex decide candidate] self parent hash is nil and vertex is not genesis, stop processing it.")
			return ProcessResult_No, nil
		}
	}

	// If self parent status is not ready yet, just return Undecided and wait for processing again
	selfParentStatus := GetVertexStatus(dagStorage, link.SelfParentHash)
	if selfParentStatus == nil || selfParentStatus.Level == 0 {
		log.W("[dag][process vertex decide candidate] cannot find self parent status, postpone it to next execution.")
		return ProcessResult_Undecided, link.SelfParentHash
	}

	vertexStatus := GetVertexStatus(dagStorage, vertexHash)
	if vertexStatus == nil {
		vertexStatus = &DagVertexStatus{}
	}

	vertexStatus.Level = selfParentStatus.Level

	if link.PeerParentHash != nil {
		peerParentStatus := GetVertexStatus(dagStorage, link.PeerParentHash)
		if peerParentStatus == nil || peerParentStatus.Level == 0 {
			log.W("[dag][process vertex decide candidate] cannot find peer parent status, postpone it to next execution.")
			return ProcessResult_Undecided, link.PeerParentHash
		}

		peerParentLink := GetVertexLink(dagStorage, link.PeerParentHash)
		if peerParentLink == nil {
			// This should not happen
			log.W("[dag][process vertex decide candidate] error while getting peer parent link.")
			return ProcessResult_No, nil
		}

		if selfParentStatus.Level < peerParentStatus.Level {

			// If this level is higher than its parent, assign this vertex as candidate
			vertexStatus.IsCandidate = true
			vertexStatus.Level = peerParentStatus.Level
		}
	}

	log.D("[dag][process vertex decide candidate] current level of the vertex:", vertexStatus.Level)

	// Calculate connection with each of the unconfirmed vertexes
	dagStorage.levelQueueUnconfirmedVertex.StartIterate()
	_, targetVertexHash := dagStorage.levelQueueUnconfirmedVertex.IterateNext()
	log.D("[dag][process vertex decide candidate] calculate connections.")

	for targetVertexHash != nil {
		log.D("[dag][process vertex decide candidate] processing vertex", GetShortenedHash(targetVertexHash))

		CalculateVertexConnection(dagStorage, vertexHash, targetVertexHash)
		_, targetVertexHash = dagStorage.levelQueueUnconfirmedVertex.IterateNext()
	}

		// Calculate the strong connection
	strongConnectionCount := 0
	for _, dagNode := range dagNodes.AllNodes() {

			candidateHash, _ := GetCandidateForNode(dagStorage, dagNode.NodeId, vertexStatus.Level, true)
			if candidateHash == nil {
				// It's possible that for given node and level there is no candidate
				continue
			}

			log.D("[dag][process vertex decide candidate] got candidate on node=", dagNode.NodeId, "level=", vertexStatus.Level, "vertex_hash=", GetShortenedHash(candidateHash))
			connection := CalculateVertexConnection(dagStorage, vertexHash, candidateHash)
			if connection == nil {
				// Data is missing, so skip this connection for now
				continue
			}

			if connection.IsConnected() && connection.GetNodeCount() >= dagNodes.GetMajorityCount() {
				log.D("[dag][process vertex decide candidate] connection from", GetShortenedHash(vertexHash), "to candidate", GetShortenedHash(candidateHash), "is strong.")
				strongConnectionCount ++
			} else {
				log.D("[dag][process vertex decide candidate] connection from", GetShortenedHash(vertexHash), "to candidate", GetShortenedHash(candidateHash), "is not strong.")
			}
	}

	log.D("[dag][process vertex decide candidate] strong connection count:", strongConnectionCount)

	if strongConnectionCount >= dagNodes.GetMajorityCount() {
		vertexStatus.Level = vertexStatus.Level + 1
		vertexStatus.IsCandidate = true
	}

	log.I("[dag][process vertex decide candidate] set status for vertex. level=", vertexStatus.Level, "isCandidate=",vertexStatus.IsCandidate, "vertex=", GetShortenedHash(vertexHash))

	// Save the vertex status
	SetVertexStatus(dagStorage, vertexHash, vertexStatus)

	// Push to queue and channel if necessary
	err := dagStorage.levelQueueUnconfirmedVertex.Push(vertexStatus.Level, vertexHash)
	if err != nil {
		// temporary fail, should retry
		log.W("[dag][process vertex decide candidate] push to level queue failed:", err.Error())
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

	log.I("[dag][processing candidate] started. candidate hash=", GetShortenedHash(nowCandidateHash))
	nowCandidateStatus := GetVertexStatus(dagStorage, nowCandidateHash)

	if nowCandidateStatus == nil {
		log.I("[dag][processing candidate] cannot find the status for candidate. stop processing.")
		return ProcessResult_No
	}

	// newQueenUpdated := ProcessResult(ProcessResult_No)

	dagStorage.levelQueueUndecidedCandidate.StartIterate()
	targetCandidateIndex, targetCandidateHash := dagStorage.levelQueueUndecidedCandidate.IterateNext()
	for targetCandidateHash != nil {

		log.D("[dag][processing candidate] target candidate:", GetShortenedHash(targetCandidateHash))
		targetCandidateStatus := GetVertexStatus(dagStorage, targetCandidateHash)
		if targetCandidateStatus == nil {
			// Status of the target candidate is wrong, should remove the target?
			dagStorage.levelQueueUndecidedCandidate.Delete(targetCandidateIndex)
			targetCandidateIndex, targetCandidateHash = dagStorage.levelQueueUndecidedCandidate.IterateNext();
			continue
		}

		if nowCandidateStatus.Level >= targetCandidateStatus.Level + 10 {

			log.I("[dag][candidate vote] it's coin round")

			// Coin round to decide Yes or No
			// TODO: just put Yes Decision for now
			SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_DecideYes)
			dagStorage.levelQueueUndecidedCandidate.Delete(targetCandidateIndex)

		} else if nowCandidateStatus.Level > targetCandidateStatus.Level + 1 {

			// Collect the vote results
			yesCount := 0
			noCount := 0
			subLevel := nowCandidateStatus.Level - 1
			log.I("[dag][processing candidate] collecting vote results in sub level=", subLevel)
			for _, node := range dagNodes.AllNodes() {

				subCandidateHash, _ := GetCandidateForNode(dagStorage, node.NodeId, subLevel, true)
				if subCandidateHash == nil {
					log.D("[dag][processing candidate] sub candidate is nil for node:", node.NodeId)
					continue
				}

				// Only collect decisions from strong connected candidates
				connection := GetVertexConnection(dagStorage, nowCandidateHash, subCandidateHash)
				if connection == nil {
					log.D("[dag][processing candidate] connection to sub candidate is nil.")
					continue
				}

				if connection.GetNodeCount() < dagNodes.GetMajorityCount() {
					log.I("[dag][processing candidate] connection is not strong, node count=", connection.GetNodeCount())
					continue
				}

				subDecision := GetCandidateDecision(dagStorage, subCandidateHash, targetCandidateHash)
				log.I("[dag][processing candidate] get candidate decision", subDecision, " from", GetShortenedHash(subCandidateHash), "to", GetShortenedHash(targetCandidateHash))
				switch subDecision {
				case CandidateDecision_No:
					noCount = noCount + 1
					break;
				case CandidateDecision_DecideNo:
					noCount = noCount + 1
					break;
				case CandidateDecision_Yes:
					yesCount = yesCount + 1
					break;
				case CandidateDecision_DecideYes:
					// This will not happen, since it's already decided by the Decision Yes
					yesCount = yesCount + 1
					break;
				case CandidateDecision_Unknown:
					break;
				}
			}

			log.I("[dag][processing candidate] yes vote count:", yesCount, "no vote count:", noCount)

			if yesCount >= dagNodes.GetMajorityCount() {

				log.I("[dag][processing candidate] Decide YES for", GetShortenedHash(targetCandidateHash))
				SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_DecideYes)

				// Change the candidate to queen
				targetCandidateStatus.IsQueenDecided = true
				targetCandidateStatus.IsQueen = true
				SetVertexStatus(dagStorage, targetCandidateHash, targetCandidateStatus)

				onQueenFound(targetCandidateHash)
				dagStorage.levelQueueUndecidedCandidate.Delete(targetCandidateIndex)
				// newQueenUpdated = ProcessResult_Yes

			} else if noCount >= dagNodes.GetMajorityCount() {

				log.I("[dag][processing candidate] Decide NO for", GetShortenedHash(targetCandidateHash))
				SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_DecideNo)

				// Change the candidate to queen
				targetCandidateStatus.IsQueenDecided = true
				targetCandidateStatus.IsQueen = false
				SetVertexStatus(dagStorage, targetCandidateHash, targetCandidateStatus)

				dagStorage.levelQueueUndecidedCandidate.Delete(targetCandidateIndex)
				// newQueenUpdated = ProcessResult_Yes

			} else if yesCount >= noCount {
				log.I("[dag][processing candidate] Decide to vote Yes for", GetShortenedHash(targetCandidateHash))
				SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_Yes)
			} else if yesCount < noCount {
				log.I("[dag][processing candidate] Decide to vote No for", GetShortenedHash(targetCandidateHash))
				SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_No)
			} else {
				log.I("[dag][processing candidate] Cannot decide to vote for", GetShortenedHash(targetCandidateHash))
				SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_Unknown)
			}

		} else if nowCandidateStatus.Level == targetCandidateStatus.Level + 1 {

			// Vote the candidate
			connection := GetVertexConnection(dagStorage, nowCandidateHash, targetCandidateHash)
			if connection != nil && connection.IsConnected() {
				log.I("[dag][processing candidate] vote Yes for vertex", GetShortenedHash(targetCandidateHash), " in sub level=", targetCandidateStatus.Level)
				SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_Yes)
			} else {
				log.I("[dag][processing candidate] vote No for vertex", GetShortenedHash(targetCandidateHash), " in sub level=", targetCandidateStatus.Level)
				SetCandidateDecision(dagStorage, nowCandidateHash, targetCandidateHash, CandidateDecision_No)
			}
		}

		targetCandidateIndex, targetCandidateHash = dagStorage.levelQueueUndecidedCandidate.IterateNext();
	}

	return ProcessResult_Yes
}

// Queen to decide whether a vertex is Accepted/Rejected
// 1. For each vertex in unconfirmedVertexQueue, use the queen to decide
// 2. If everything goes well, return Yes
// 3. If wrong happen, return No. The upper will re-process the queen later
// 4. Collect all the queens on the level, if there is any candidates not decided to be queen, stop processing
// 5. Make sure the queens on the level is more than super majority, then keep processing
// 6. For each unconfirmed vertex, if it can be connected by all the queens, mark it as Accepted
func ProcessQueenDecision(dagStorage *DagStorage, dagNodes *DagNodes, queenHash []byte,
	vertexConfirmer func (vertexHash []byte, result VertexConfirmResult)) ProcessResult {

	log.I("[dag][queen decision] start.")
	queenStatus := GetVertexStatus(dagStorage, queenHash)
	if queenStatus == nil {
		log.W("[dag][queen decision] queen status is nil.")
		return ProcessResult_No
	}

	allCandidatesDecided := true
	allQueensInLevel := make([][]byte, 0)
	for _, node := range dagNodes.AllNodes() {

		candidateHash, _ := GetCandidateForNode(dagStorage, node.NodeId, queenStatus.Level, true)
		if candidateHash == nil {
			log.I("[dag][queen decision] no candidate for node", node.NodeId, "in level", queenStatus.Level)
			continue
		}

		candidateStatus := GetVertexStatus(dagStorage, candidateHash)
		if candidateStatus == nil || !candidateStatus.IsQueenDecided {
			log.I("[dag][queen decision] isQueenDecided is false for candidate, stop processing.")
			allCandidatesDecided = false
			break
		}

		if candidateStatus.IsQueen {
			log.I("[dag][queen decision] adding queen", GetShortenedHash(candidateHash), "to allQueensInLevel")
			allQueensInLevel = append(allQueensInLevel, candidateHash)
		}
	}

	if !allCandidatesDecided || len(allQueensInLevel) < dagNodes.GetMajorityCount() {

		// Not all candidates on this level decided is queen or not, or there is no queen in this level,
		// just pass this and do nothing
		log.I("[dag][queen decision] not enough queen or some are not decided to do the decision, stop processing.")
		return ProcessResult_Yes
	}

	// Start iterating all the vertexes, and using the AllQueensInLevel to decide accept/reject it
	dagStorage.levelQueueUnconfirmedVertex.StartIterate()
	targetVertexIndex, targetVertexHash := dagStorage.levelQueueUnconfirmedVertex.IterateNext()
	log.I("[dag][queen decision] start to confirm vertexes.")

	for targetVertexHash != nil {
		log.I("[dag][queen decision] processing vertex", GetShortenedHash(targetVertexHash))

		hasAllConnection := true
		for _, iQueenHash := range allQueensInLevel {

			connection := GetVertexConnection(dagStorage, iQueenHash, targetVertexHash)
			if connection == nil || !connection.IsConnected() {
				log.I("[dag][queen decision] queen", GetShortenedHash(iQueenHash), "is not connected to vertex, stop processing")
				hasAllConnection = false
				break
			}
		}

		if hasAllConnection {

			log.I("[dag][queen decision] all queens are connected to this vertex, confirm as Accepted.")
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
		log.I("[dag] got self genesis vertex=", GetShortenedHash(genesisVertexHash))
		return false, genesisVertexHash
	}

	// Create if missing
	vertex, err := CreateGenesisVertex(dagStorage, node)
	if err != nil {
		return false, nil
	}

	SetGenesisVertex(dagStorage, node.NodeId, vertex.Hash)
	SetNodeLatestVertex(dagStorage, node.NodeId, vertex.Hash)

	log.I("[dag] created self genesis vertex=", GetShortenedHash(vertex.Hash))
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

	// Update the knowledge of the vertexes
	SetNodeSyncVertex(dagStorage, nodeId, vertexHash)

	return true
}

//
// CalculateVertexConnection:
// This should always return a non-null value:
//
func CalculateVertexConnection(dagStorage *DagStorage, vertexHash []byte, targetVertexHash []byte) *DagVertexConnection {

	log.D("[dag][calculate connection] start calculating from", GetShortenedHash(vertexHash), "to", GetShortenedHash(targetVertexHash))
	vertexLink := GetVertexLink(dagStorage, vertexHash)
	if vertexLink == nil {
		// Something wrong unexpected, just return nil
		log.W("[dag][calculate connection] cannot find the vertex link, cancel it.")
		return nil
	}

	existingConnection := GetOrDefaultVertexConnection(dagStorage, vertexHash, targetVertexHash, vertexLink.NodeId)
	if existingConnection != nil {
		log.D("[dag][calculate connection] found existing connection, node list=", existingConnection.NodeIdList)
		return existingConnection
	}

	nodeIdList := make([]uint64, 0)

	// Consider the genesis vertex which self parent is nil, so adding this check
	if vertexLink.SelfParentHash != nil {
		selfParentConnection := GetOrDefaultVertexConnection(dagStorage, vertexLink.SelfParentHash, targetVertexHash, vertexLink.NodeId)
		if selfParentConnection == nil {
			// Self parent is not connected to target, and never calculated
			// TODO: should we consider it's not done yet?
			log.D("[dag][calculate connection] self parent connection is not ready, consider it's not connected.")
		} else {
			nodeIdList = selfParentConnection.NodeIdList
		}
	} else {
		log.W("[dag][calculate connection] self parent hash is nil.")
	}

	if vertexLink.PeerParentHash != nil {

		peerParentLink := GetVertexLink(dagStorage, vertexLink.PeerParentHash)
		if peerParentLink != nil {
			peerParentConnection := GetOrDefaultVertexConnection(dagStorage, vertexLink.PeerParentHash, targetVertexHash, peerParentLink.NodeId)

			if peerParentConnection == nil {
				// Peer parent connection is not ready yet, and never calculated
				// TODO: should we consider it's not done yet?
				log.D("[dag][calculate connection] peer parent connection is not ready, consider it's not connected.")
			} else {
				nodeIdList = MergeUint64Array(nodeIdList, peerParentConnection.NodeIdList)
			}
		}
	} else {
		log.D("[dag][calculate connection] peer parent hash is nil.")
	}

	connectionResult := NewDagVertexConnection()
	log.D("[dag][calculate connection] composing connection result.")

	// merge with current node only if the previous merged list is not empty
	if len(nodeIdList) > 0 {
		log.D("[dag][calculate connection] merging the node list with current node id. node list=", nodeIdList, "current node id=", vertexLink.NodeId)
		connectionResult.NodeIdList = MergeUint64Array(nodeIdList, []uint64{ vertexLink.NodeId })
	}

	// Save it into storage
	SetVertexConnection(dagStorage, vertexHash, targetVertexHash, connectionResult)
	log.D("[dag][calculate connection] successfully saved connection result: from", GetShortenedHash(vertexHash), "to", GetShortenedHash(targetVertexHash), ". node list=", connectionResult.NodeIdList)

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

		if index1 >= len(array1) {
			result = append(result, array2[index2])
			index2 ++
		} else if index2 >= len(array2) {
			result = append(result, array1[index1])
			index1 ++
		} else if array2[index2] > array1[index1] {
			result = append(result, array1[index1])
			index1 ++
		} else if array2[index2] < array1[index1] {
			result = append(result, array2[index2])
			index2 ++
		} else {
			// They are equal, just pick one value
			result = append(result, array1[index1])
			index1 ++
			index2 ++
		}
	}

	return result
}

func MergeVertexes(vertexList[]*DagVertex, vertex *DagVertex) (resultList []*DagVertex) {

	if vertex == nil && vertexList == nil {
		return nil
	}

	if vertex == nil {
		return vertexList
	}

	if vertexList == nil {
		return []*DagVertex { vertex }
	}

	for _, v := range vertexList {
		if bytes.Equal(v.Hash, vertex.Hash) {
			// If found the vertex, just return the origin list
			return vertexList
		}
	}

	return append(vertexList, vertex)
}

func GetShortenedHash(hash []byte) string {

	if hash == nil {
		return "nil"
	}

	hashString := hex.EncodeToString(hash)

	return hashString[:5] + "..." + hashString[len(hashString)-2:]
}