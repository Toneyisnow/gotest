package dag

func NewDagVertexConnection() *DagVertexConnection {

	connection := &DagVertexConnection{}
	connection.NodeIdList = make([]uint64, 0)

	return connection
}

func (this *DagVertexConnection) IsConnected() bool {

	return len(this.NodeIdList) > 0
}

func (this *DagVertexConnection) GetNodeCount() int {
	return len(this.NodeIdList)
}