package services

import (
	"networking/pb"
	"time"
)

type NodeClientManager struct {

	_clientList map[string]*NodeClient
	_clientLastVisitStamp map[string]time.Time

	_maxActiveClient int

	//// _mainMessageQueue chan *pb.NetMessage

	_clientMessageQueue map[string]chan *pb.NetMessage

	_nodeConfig *NodeConfig
}

func (this *NodeClientManager) Initialize (maxActiveClient int) {

	//// this._mainMessageQueue = mainQueue
	this._maxActiveClient = maxActiveClient

	this._clientList = make(map[string]*NodeClient)
	this._clientLastVisitStamp = make(map[string]time.Time)

	this._nodeConfig = LoadConfigFromFile()
}

func (this *NodeClientManager) SendToNode(nodeId string, message *pb.NetMessage) {

	// nodeInfo := this._nodeConfig.GetPeerById(nodeId)

	nodeClient, exists := this._clientList[nodeId]
	if (!exists) {
		nodeClient = new(NodeClient)
		this._clientMessageQueue[nodeId] = make(chan *pb.NetMessage)

		nodeClient.Initialize(nodeId, this._clientMessageQueue[nodeId])
		go nodeClient.Start()

		this._clientList[nodeId] = nodeClient
		this._clientLastVisitStamp[nodeId] = time.Now()
	}

	msgQueue := this._clientMessageQueue[nodeId]
	if (msgQueue == nil) {
		// Error
		return
	}

	msgQueue <- message
}

