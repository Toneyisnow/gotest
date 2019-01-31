package contract

import (
	"../dag"
	"fmt"
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

	handler := ContractPayloadHandler{}
	this.dagEngine = dag.NewDagEngine(config, &handler)

	log.I("[contract] starting engine...")
	go this.dagEngine.Start()

	// Wait for the
	log.I("[contract] waiting engine to be online...")
	time.Sleep(3 * time.Second)
	for {
		if this.dagEngine.IsOnline() {
			break
		}

		time.Sleep(time.Second)
	}
	log.I("[contract] engine is online. contract executor initialized.")
}

func (this *ContractExecutor) CreateWallet(walletName string) (payloadId PayloadId, privateKey []byte) {

	if !this.dagEngine.IsOnline() {
		log.W("[contract] create wallet: dag engine is not online now, cannot proceed.")
		return nil, nil
	}

	payload, privateKey := NewCreateWalletPayload(walletName)
	err := SavePayload(this.contractStorage, payload)
	if err != nil {

	}

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


	// Validate the asset type and Id


	payload := NewTransactionPayload(nil, nil, nil)
	err := SavePayload(this.contractStorage, payload)
	if err != nil {

	}

	return payload.PayloadId
}

func (this *ContractExecutor) QueryPayload(payloadId PayloadId) *ContractPayload {

	if payloadId == nil {
		return nil
	}

	return GetPayload(this.contractStorage, payloadId)
}

