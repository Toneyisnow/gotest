package types

import "common"

type SWNode struct {
	NodeId string
}

func CreateNode(info *common.NodeInfo) *SWNode {

	node := new (SWNode)
	node.NodeId = info.Identifier

	return node
}

