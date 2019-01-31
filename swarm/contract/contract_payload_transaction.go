package contract

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/smartswarm/go/log"
	"time"
)

func NewTransactionPayload(sourceWallet *ContractWallet, destinationWallet *ContractWallet, asset *ContractAssset) *ContractPayload {

	if sourceWallet == nil || destinationWallet == nil || asset == nil {
		log.W("[contract][new transaction payload] cannot create, nil parameter.")
		return nil
	}

	payload := new(ContractPayload)
	payload.PayloadId = GeneratePayloadId()
	payload.CreatedTime, _ = ptypes.TimestampProto(time.Now())
	payload.PayloadStatus = ContractPayloadStatus_Initialized
	payload.PayloadType = ContractPayloadType_Transaction

	transaction := new(TransactionPayload)
	transaction.Asset = asset
	transaction.SourceWalletAddress = sourceWallet.Address
	transaction.DestinationWalletAddress = destinationWallet.Address
	transaction.TransactionId = GeneratePayloadId()

	payload.Data = &ContractPayload_TransactionPayload{TransactionPayload: transaction}

	return payload
}

func (this *TransactionPayload) Validate() bool {

	if this.SourceWalletAddress == nil {
		return false
	}

	if this.DestinationWalletAddress == nil {
		return false
	}

	if this.TransactionId == nil {
		return false
	}

	if this.Asset == nil {
		return false
	}

	// Check the source wallet balance

	return true
}