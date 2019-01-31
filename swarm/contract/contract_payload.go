package contract

import (
	"errors"
	"github.com/google/uuid"
)

func (this *ContractPayload) Validate(storage *ContractStorage) bool {

	if this == nil {
		return false
	}

	switch(this.PayloadType) {
	case ContractPayloadType_CreateWallet:
		if this.GetCreateWalletPayload() != nil {
			return this.GetCreateWalletPayload().Validate(storage)
		}
		break;
	case ContractPayloadType_Transaction:
		if this.GetTransactionPayload() != nil {
			return this.GetTransactionPayload().Validate(storage)
		}
		break;
	}

	return false
}

func (this *ContractPayload) Process(storage *ContractStorage) error {

	if this == nil {
		return errors.New("[contract][process] cannot process since payload is nil.")
	}

	switch(this.PayloadType) {
	case ContractPayloadType_CreateWallet:
		if this.GetData() != nil {
			return this.GetCreateWalletPayload().Process(storage)
		}
		break;
	case ContractPayloadType_Transaction:
		if this.GetTransactionPayload() != nil {
			return this.GetTransactionPayload().Process(storage)
		}
		break;
	}

	return errors.New("[contract][process] payload type is wrong:" + string(this.PayloadType))
}

func (this *ContractPayload) GetReceipt() *ContractReceipt {

	if this.PayloadStatus != ContractPayloadStatus_Accepted &&
		this.PayloadStatus != ContractPayloadStatus_Rejected &&
		this.PayloadStatus != ContractPayloadStatus_Rejected_By_Dag {

		// The receipt is not ready yet
		return nil
	}

	receipt := &ContractReceipt{}
	receipt.PayloadId = this.PayloadId
	receipt.CreatedTime = this.CreatedTime
	receipt.ConfirmedTime = this.ConfirmedTime

	if this.PayloadStatus == ContractPayloadStatus_Accepted {
		receipt.Status = ContractReceiptStatus_Accepted
	} else {
		receipt.Status = ContractReceiptStatus_Rejected
	}

	switch this.PayloadType {
	case ContractPayloadType_CreateWallet:
		if this.GetCreateWalletPayload() != nil {
			receipt.CreatorAddress = this.GetCreateWalletPayload().WalletAddress
		}
		break;
	case ContractPayloadType_Transaction:
		if this.GetTransactionPayload() != nil {
			receipt.CreatorAddress = this.GetTransactionPayload().SourceWalletAddress
		}
		break;
	}

	return receipt
}

func GeneratePayloadId() []byte {

	uuid, err := uuid.NewUUID()
	if err != nil {
		return nil
	}

	bytes, err := uuid.MarshalBinary()
	if err != nil {
		return nil
	}

	return bytes
}

