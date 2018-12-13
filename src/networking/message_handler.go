package networking

import (
	"common"
	"components/types"
	"log"
	"networking/pb"
)

type MessageHandler struct {

}

func ComposeNetMessage(events []*types.SWEvent) *pb.NetMessage {

	log.Printf("events len: %d", len(events))

	message := new(pb.NetMessage)

	// Add signature from current node as OwnerNode
	config := common.LoadConfigFromFile()
	message.OwnerId = config.Self.Identifier
	message.MessageType = pb.NetMessageType_SendEvents
	message.SendEventMessage = new(pb.SendEventMessage)

	for _, eve := range events {

		if (eve == nil) {
			log.Println("eve is nil")
		}

		netEvent := ConvertEvent(eve)
		message.SendEventMessage.Events = append(message.SendEventMessage.Events, netEvent)
	}

	return message
}

func ConvertEvent(event *types.SWEvent) *pb.NetEvent {

	netEvent := new(pb.NetEvent)

	if (event.OwnerNode == nil) {
		log.Println("event.OwnerNode is nil")
	}

	netEvent.OwnerId = event.OwnerNode.NodeId
	netEvent.EventId = event.EventId
	for _, tran := range event.TransactionList {

		netTransaction := new(pb.NetTransaction)
		netTransaction.InAccount = tran.SubjectAccount.Address
		netTransaction.OutAccount = tran.ObjectAccount.Address

		netEvent.Transactions = append(netEvent.Transactions, netTransaction)
	}

	log.Printf("netEvent Transactions len: %d", len(netEvent.Transactions))

	return netEvent
}

