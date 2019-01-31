package contract

import (
	"errors"
	"github.com/golang/protobuf/ptypes"
	"github.com/smartswarm/go/log"
	"time"
)

func NewTransactionPayload(sourceWallet *ContractWallet, destinationWallet *ContractWallet, asset *ContractAsset) *ContractPayload {

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

func (this *TransactionPayload) Validate(storage *ContractStorage) bool {

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
	sourceWallet := GetWallet(storage, this.SourceWalletAddress)
	if sourceWallet == nil {
		log.W("[contract] cannot find source wallet address.")
		return false
	}

	destinationWallet := GetWallet(storage, this.DestinationWalletAddress)
	if destinationWallet == nil {
		log.W("[contract] cannot find destination wallet address.")
		return false
	}

	if !sourceWallet.HasEnough(this.Asset) {
		return false
	}

	return true
}

func (this *TransactionPayload) Process(storage *ContractStorage) error {

	log.I("[contract] processing transaction.")

	sourceWallet := GetWallet(storage, this.SourceWalletAddress)
	destinationWallet := GetWallet(storage, this.DestinationWalletAddress)
	if sourceWallet == nil || destinationWallet == nil {
		return errors.New("[contract] error while processing.")
	}

	// Should make transaction operation here
	sourceWallet.RemoveAsset(this.Asset)
	destinationWallet.AddAsset(this.Asset)

	SaveWallet(storage, sourceWallet)
	SaveWallet(storage, destinationWallet)

	return nil
}