package contract

import (
	"errors"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/smartswarm/core/crypto/secp256k1"
	"github.com/smartswarm/go/log"
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

	return payload, privKey;
}

func (this *CreateWalletPayload) Validate(storage *ContractStorage) bool {

	if this.WalletAddress == nil {
		return false
	}

	return true
}

func (this *CreateWalletPayload) Process(storage *ContractStorage) error {

	log.I("[contract] processing create wallet.")
	// Actually create the wallet in storage

	wallet := NewContractWallet(this.WalletName, this.WalletAddress)
	if wallet == nil {
		return errors.New("create new wallet failed")
	}

	walletBytes, err := proto.Marshal(wallet)
	if err != nil {
		return errors.New("create new wallet failed")
	}

	err = storage.tableWallet.InsertOrUpdate(wallet.Address, walletBytes)
	if err == nil {
		log.I("[contract] create new wallet succeeded.")
	}

	return err
}