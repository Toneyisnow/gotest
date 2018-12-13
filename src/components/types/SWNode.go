package types

import "services"

type SWNode struct {
	NodeId string
}

func CreateNode(info *services.NodeInfo) *SWNode {

	node := new (SWNode)
	node.NodeId = info.Identifier

	return node
}

