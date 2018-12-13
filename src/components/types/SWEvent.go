package types

import "log"

type SWEvent struct {

	EventId string
	OwnerNode *SWNode
	TransactionList []*SWTransaction
}

func NewEvent(oNode *SWNode) *SWEvent {

	eve := new(SWEvent)

	eve.EventId = "101"

	if (oNode == nil) {
		log.Println("oNode is nil.")
	}

	eve.OwnerNode = oNode
	eve.TransactionList = []*SWTransaction {}

	return eve
}

func (this *SWEvent) AddTransaction(transaction *SWTransaction) {

	log.Printf("this.TransactionList len: %d", len(this.TransactionList))

	this.TransactionList = append(this.TransactionList, transaction)
}