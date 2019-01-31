package app

import (
	"../contract"
	"encoding/hex"
	"github.com/smartswarm/go/log"
	"time"
)


// Instruction:
// 1. Build the swarm folder, generate <execution_file>
// 2. Create a folder <node_folder_1>, copy <execution_file> and swarm-config-1.json to the folder, run:
//        <execution_file> swarm-config-1.json
// 3. Create folder <node_folder_2>, <node_folder_3>, and do the same as step 2
//
func RunSample() {

	contractExecutor := contract.NewContractExecutor()

	// Executor will wait to be connected here
	contractExecutor.Initialize()

	// Create wallet, you should save the private key.
	// For testing purpose, every wallet will be given 100 coins after created
	payloadIdA, privateKeyA := contractExecutor.CreateWallet("ToneyWalletA")
	log.I("payloadA id=", hex.EncodeToString(payloadIdA), "private key=", hex.EncodeToString(privateKeyA))

	payloadIdB, privateKeyB := contractExecutor.CreateWallet("ToneyWalletB")
	log.I("payloadB id=", hex.EncodeToString(payloadIdB), "private key=", hex.EncodeToString(privateKeyB))

	var receiptA *contract.ContractReceipt
	var receiptB *contract.ContractReceipt

	for {
		payloadA := contractExecutor.QueryPayload(payloadIdA)
		payloadB := contractExecutor.QueryPayload(payloadIdB)

		if payloadA != nil && payloadA.GetReceipt() != nil && payloadB != nil && payloadB.GetReceipt()!= nil {

			receiptA = payloadA.GetReceipt()
			receiptB = payloadB.GetReceipt()
			break
		}
		time.Sleep(time.Second)
	}

	log.I("[app] The payloadA and payloadB are confirmed.")
	if receiptA.Status != contract.ContractReceiptStatus_Accepted && receiptB.Status != contract.ContractReceiptStatus_Accepted {
		log.W("[app] receip status is not correct.")
		return
	}

	payloadIdC := contractExecutor.ComposeTransaction(receiptA.CreatorAddress, receiptB.CreatorAddress, contract.ContractAssetType_Coin, 0, 20)
	var receiptC *contract.ContractReceipt
	for {
		payloadC := contractExecutor.QueryPayload(payloadIdC)

		if payloadC != nil && payloadC.GetReceipt() != nil {

			receiptC = payloadC.GetReceipt()
			break
		}
		time.Sleep(time.Second)
	}

	log.I("[app] transaction result:", receiptC.Status)
}

