package network

import (
	"encoding/json"
)

type HGMessage struct {

	OwnerId string		// Id of the node
	Signature string	// Signature from the node

	Events []HGEvent
}


func ReadMessageFromJson(jsonString string) *HGMessage {

	var message = new(HGMessage)
	json.Unmarshal([]byte(jsonString), message)

	return message
}


