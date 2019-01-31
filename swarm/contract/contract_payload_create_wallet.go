package contract

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/smartswarm/core/crypto/secp256k1"
	"time"
)

func NewCreateWalletPayload(walletName string) (result *ContractPayload, privateKey []byte) {

	pubKey, privKey := secp256k1.GenerateKeyPair()

	payload := new(ContractPayload)
	payload.PayloadId = GeneratePayloadId()
	payload.CreatedTime, _ = ptypes.TimestampProto(time.Now())
	payload.PayloadStatus = ContractPayloadStatus_Initialized
	payload.PayloadType = ContractPayloadType_CreateWallet

	createWallet := new(CreateWalletPayload)
	createWallet.WalletName = walletName
	createWallet.WalletAddress = pubKey
	createWallet.WalletPublicKey = pubKey

	payload.Data = &ContractPayload_CreateWalletPayload{CreateWalletPayload: createWallet}

	return result, privKey;
}

func (this *CreateWalletPayload) Validate() bool {

	return true
}

func (this *CreateWalletPayload) Process() error {

	// Actually create the wallet in storage

	return nil
}