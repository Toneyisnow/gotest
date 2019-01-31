package app

import (
	"../contract"
	"encoding/hex"
	"github.com/smartswarm/go/log"
)

func RunSample() {

	contractExecutor := contract.NewContractExecutor()

	contractExecutor.Initialize()

	// Create wallet, you should save the private key
	payloadId, privateKey := contractExecutor.CreateWallet("ToneyWallet")
	log.I("payload id=", hex.EncodeToString(payloadId), "private key=", hex.EncodeToString(privateKey))

	for {
		payload := contractExecutor.QueryPayload(payloadId)
		if payload.PayloadStatus == contract.ContractPayloadStatus_Accepted {
			break
		}
	}
}

