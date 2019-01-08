package dag

type ProcessResult int

const (
	ProcessResult_No        = 0
	ProcessResult_Yes       = 1
	ProcessResult_Undecided = 2
)

// Choose one DagNode to send vertexes that it might not know, return nil if no need to send
func SelectPeerNodeToSendVertex(storage *DagStorage, nodes *DagNodes) (node *DagNode) {

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
// 3. Put the vertex into tableVertex, queueFreshVertex, queueUnconfirmedVertex
func ProcessIncomingVertex(storage *DagStorage, incomingVertex *DagVertex) ProcessResult {

	return ProcessResult_Yes
}

// ProcessVertexAndDecideCandidate:
// 1. Calculate the Sees() for the freshVertex vs. each of the vertexes in candidates
// 2. If the Sees() result is greater than majority of nodes for majority of candidates, then mark this vertex
// to be level+1
// 3. If this is the first vertex in the level for this node, mark this vertex as candidate and return Yes, otherwise No
// 4. Save this vertex into tableLatestVertex, tableCandidate
func ProcessVertexAndDecideCandidate(storage *DagStorage, freshVertexHash []byte) ProcessResult {

	return ProcessResult_No
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
