package objectmodels

import (
	"encoding/json"
	"fmt"
)

type HGMessage struct {

	OwnerId string		// Id of the node
	Signature string	// Signature from the node
	
	Events []HGEvent
}

type HGEvent struct {

	Signature string
	SenderNode string
	ReceiverNode string
	Hash string
	SelfParentHash string
	PeerParentHash string
	
	Transactions []HGTransaction
	
}

type HGTransaction struct {

	Hash string
	Event string

	FromAddress string
	ToAddress string
	Amount uint64
}

type HGNode struct {

	
	HostUrl string `json:"HostUrl"`
	Identifier string `json:"Identifier"`
	PublickKey string `json:"PublicKey"`
	
}

func ReadMessageFromJson(jsonString string) *HGMessage {

	var message = new(HGMessage)
	json.Unmarshal([]byte(jsonString), message)

	return message
}

func ReadNodeFromJson(jsonString string) *HGNode {

	var node = new(HGNode)

	json.Unmarshal([]byte(jsonString), node)
	fmt.Println("Node: " + node.HostUrl)


	return node
}