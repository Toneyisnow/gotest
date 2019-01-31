package contract

import (
	"../common/log"
	"../dag"
	"../storage"
	"github.com/gogo/protobuf/proto"
)

type ContractPayloadHandler struct {

	chanReceivedPayloads *storage.RocksChannel
	contractStorage *ContractStorage
}

func NewContractPayloadHandler(contractStorage *ContractStorage) *ContractPayloadHandler {

	handler := &ContractPayloadHandler{}
	handler.chanReceivedPayloads = contractStorage.chanReceivedPayload
	handler.contractStorage = contractStorage

	return handler
}

func (this *ContractPayloadHandler) OnPayloadSubmitted(data dag.PayloadData) {

	log.I("[contract] on payload submitted: ", data)
}

func (this *ContractPayloadHandler) OnPayloadAccepted(data dag.PayloadData) {

	log.I("[contract] on payload accepted: ", data)
	this.chanReceivedPayloads.Push(data)
}

func (this *ContractPayloadHandler) OnPayloadRejected(data dag.PayloadData) {

	log.I("[contract] on payload rejected: ", data)
}

func (this *ContractPayloadHandler) ProcessPayloadResult(data []byte) {

	result := &ContractPayloadDagResult{}
	err := proto.Unmarshal(data, result)
	if err != nil {

	}

	if result.PayloadData == nil {

		log.W("[contract][process payload result] payload data is nil.")
		return
	}

	payload := &ContractPayload{}
	err = proto.Unmarshal(result.PayloadData, payload)

	if result.ResultType == ContractPayloadDagResultType_DagAccepted {

		if payload != nil && payload.Validate() {
			payload.Process()
			payload.PayloadStatus = ContractPayloadStatus_Accepted
		} else {
			payload.PayloadStatus = ContractPayloadStatus_Rejected
		}
		err = SavePayload(this.contractStorage, payload)
	} else {
		payload.PayloadStatus = ContractPayloadStatus_Rejected_By_Dag
		err = SavePayload(this.contractStorage, payload)
	}
}