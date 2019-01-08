package dag

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

