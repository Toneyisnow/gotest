package dag

import (
	"github.com/golang/protobuf/ptypes"
	"time"
)

func GenerateNewVertex(peerParent *DagVertex) (vertex *DagVertex, err error) {

	vertex = new(DagVertex)

	vertex.CreatorNodeId = 1001

	content := new(DagVertexContent)
	content.TimeStamp, _ = ptypes.TimestampProto(time.Now())

	// Mutex to read from storage
	content.SelfParentHash = ""
	content.PeerParentHash = ""
	content.Data = append(content.Data, []byte("asdadsf"))
	vertex.Content = content

	// Save to database

	err = nil
	return
}

func FindPossibleUnknownVertexesForNode(node *DagNode) (resultList []*DagVertex, err error) {

	resultList = make([]*DagVertex, 0)

	resultList = append(resultList, nil)

	err = nil
	return
}

func (this *DagVertex) SaveToStorage() (err error) {

	// vertex = new(DagVertex)


	err = nil
	return

}

