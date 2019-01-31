package contract

import (
	"../dag"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/smartswarm/go/log"
	"os"
	"time"
)

type PayloadId []byte

type ContractExecutor struct {

	contractStorage *ContractStorage

	dagEngine *dag.DagEngine
}

func NewContractExecutor() *ContractExecutor {

	executor := &ContractExecutor{}


	return executor
}

func (this *ContractExecutor) Initialize() {

	if len(os.Args) < 2 {
		fmt.Println("Missing config file in command. Usage: [_exe_] <ConfigFile.json>")
		return
	}

	log.I("[contract] initializing...")

	configFileName := string(os.Args[1])
	config := dag.LoadConfigFromJsonFile(configFileName)

	this.contractStorage = ContractStorageGetInstance(config.StorageLocation)

	handler := NewContractPayloadHandler(this.contractStorage)
	this.dagEngine = dag.NewDagEngine(config, handler)


	log.I("[contract] starting engine...")
	go this.dagEngine.Start()

	// Wait for the
	log.I("[contract] waiting for engine to be online...")
	time.Sleep(3 * time.Second)
	for {
		if this.dagEngine.IsOnline() {
			break
		}

		time.Sleep(time.Second)
	}

	log.I("[contract] engine is online. contract executor initialized.")

	handler.Start()

	// Submit dumb payloads
	go func() {

		for {

			dumb := NewDumbPayload()
			this.submitPayload(dumb)

			time.Sleep(200 * time.Millisecond)
		}
	}()
}

func (this *ContractExecutor) CreateWallet(walletName string) (payloadId PayloadId, privateKey []byte) {

	if !this.dagEngine.IsOnline() {
		log.W("[contract] create wallet: dag engine is not online now, cannot proceed.")
		return nil, nil
	}

	payload, privateKey := NewCreateWalletPayload(walletName)
	if payload == nil {
		log.W("[contract][create wallet] creating payload failed")
		return nil, nil
	}

	err := SavePayload(this.contractStorage, payload)
	if err != nil {

	}

	err = this.submitPayload(payload)

	payloadId = payload.PayloadId
	return
}

func (this *ContractExecutor) ComposeTransaction(sourceWalletAddress []byte, destinationWalletAddress []byte,
	assetType ContractAssetType, assetId uint32, amount uint64) PayloadId {

	if !this.dagEngine.IsOnline() {
		log.W("[contract] create wallet: dag engine is not online now, cannot proceed.")
		return nil
	}

	// Validate the wallet addresses
	sourceWallet := GetWallet(this.contractStorage, sourceWalletAddress)
	if sourceWallet == nil {
		log.W("[contract] cannot find source wallet address.")
		return nil
	}

	destinationWallet := GetWallet(this.contractStorage, destinationWalletAddress)
	if destinationWallet == nil {
		log.W("[contract] cannot find destination wallet address.")
		return nil
	}

	// Validate the asset type and Id
	asset := &ContractAsset{ AssetType:assetType, AssetId: assetId, Amount: amount}

	payload := NewTransactionPayload(sourceWallet, destinationWallet, asset)
	if payload == nil {
		log.W("[contract][compose transaction] creating payload failed")
		return nil
	}

	err := SavePayload(this.contractStorage, payload)
	if err != nil {

	}

	err = this.submitPayload(payload)

	return payload.PayloadId
}

func (this *ContractExecutor) QueryPayload(payloadId PayloadId) *ContractPayload {

	if payloadId == nil {
		return nil
	}

	return GetPayload(this.contractStorage, payloadId)
}

func (this *ContractExecutor) submitPayload(payload *ContractPayload) error {

	payloadBytes, err := proto.Marshal(payload)
	err = this.dagEngine.SubmitPayload(payloadBytes)

	return err
}