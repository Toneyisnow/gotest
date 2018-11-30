package network



type HGEvent struct {

	Signature string
	SenderNode string
	ReceiverNode string
	Hash string
	SelfParentHash string
	PeerParentHash string

	Transactions []HGTransaction

}
