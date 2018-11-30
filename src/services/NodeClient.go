package services

type NodeClient struct {

	config *NodeConfig
}

func (this *NodeClient) Initialize() {

	this.config = LoadConfigFromFile()
}

func (this *NodeClient) Connect(toServer *NodeConfig) {

}
