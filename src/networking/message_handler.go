package networking

import (
	"components/types"
	"networking/pb"
	"services"
)

type MessageHandler struct {

}

func ComposeNetMessage(events []*types.SWEvent) *pb.NetMessage {

	message := new(pb.NetMessage)

	// Add signature from current node as OwnerNode
	config := services.LoadConfigFromFile()
	message.OwnerId = config.Self.Identifier
	message.MessageType = pb.NetMessageType_SendEvents
	message.SendEventMessage = new(pb.SendEventMessage)

	for _, eve := range events {
		netEvent := ConvertEvent(eve)
		message.SendEventMessage.Events = append(message.SendEventMessage.Events, netEvent)
	}

	return message
}

func ConvertEvent(event *types.SWEvent) *pb.NetEvent {

	netEvent := new(pb.NetEvent)
	netEvent.OwnerId = event.OwnerNode.NodeId
	netEvent.EventId = event.EventId
	netEvent.Transactions = make([]*pb.NetTransaction, len(event.TransactionList))
	for _, tran := range event.TransactionList {

		netTransaction := new(pb.NetTransaction)
		netTransaction.InAccount = tran.SubjectAccount.Address
		netTransaction.OutAccount = tran.ObjectAccount.Address

		netEvent.Transactions = append(netEvent.Transactions, netTransaction)
	}

	return netEvent
}

