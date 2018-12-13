package types


type SWEvent struct {

	EventId string
	OwnerNode *SWNode
	TransactionList []*SWTransaction
}

func NewEvent(oNode *SWNode) *SWEvent {

	eve := new(SWEvent)

	eve.EventId = "101"
	eve.OwnerNode = oNode
	eve.TransactionList = make([]*SWTransaction, 10)

	return eve
}

func (this *SWEvent) AddTransaction(transaction *SWTransaction) {

	this.TransactionList = append(this.TransactionList, transaction)
}