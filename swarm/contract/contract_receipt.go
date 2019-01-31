package contract

import "github.com/golang/protobuf/ptypes/timestamp"

type ContractReceiptStatus int
const (
	ContractReceiptStatus_Accepted = 1
	ContractReceiptStatus_Rejected = 2
)

type ContractReceipt struct {

	PayloadId []byte
	Status ContractReceiptStatus
	CreatorAddress []byte
	CreatedTime *timestamp.Timestamp
	ConfirmedTime *timestamp.Timestamp
}





