package contract

import (
	"github.com/google/uuid"
)

func (this *ContractPayload) Validate() bool {

	if this == nil {
		return false
	}

	switch(this.PayloadType) {
	case ContractPayloadType_CreateWallet:
		if this.GetCreateWalletPayload() != nil {
			return this.GetCreateWalletPayload().Validate()
		}
		break;
	case ContractPayloadType_Transaction:
		if this.GetTransactionPayload() != nil {
			return this.GetTransactionPayload().Validate()
		}
		break;
	}

}

func (this *ContractPayload) Process() bool {

	if this == nil {
		return false
	}

	switch(this.PayloadType) {
	case ContractPayloadType_CreateWallet:
		if this.GetData() != nil {
			return this.GetCreateWalletPayload().Process()
		}
		break;
	case ContractPayloadType_Transaction:
		if this.GetTransactionPayload() != nil {
			return this.GetTransactionPayload().Process()
		}
		break;
	}
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

