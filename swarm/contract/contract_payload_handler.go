package contract

import (
	"../common/log"
	"../dag"
	"../storage"
	"encoding/hex"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"time"
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

func (this *ContractPayloadHandler) Start() {

	go this.contractStorage.chanReceivedPayload.Listen(this.ProcessPayloadResult)
	this.contractStorage.chanReceivedPayload.Reload()
}

func (this *ContractPayloadHandler) OnPayloadSubmitted(data dag.PayloadData) {

	log.I("[contract] on payload submitted. data length=", len(data))

	if data != nil {
		payload := &ContractPayload{}
		proto.Unmarshal(data, payload)
		if payload != nil {
			payload.SubmittedTime, _ = ptypes.TimestampProto(time.Now())
			SavePayload(this.contractStorage, payload)
		}
	}
}

func (this *ContractPayloadHandler) OnPayloadAccepted(data dag.PayloadData) {

	log.I("[contract] on payload accepted. data length=", len(data))
	if data != nil {

		result := &ContractPayloadDagResult{ ResultType:ContractPayloadDagResultType_DagAccepted, PayloadData:data }
		resultBytes, err := proto.Marshal(result)

		if err == nil {
			this.chanReceivedPayloads.Push(resultBytes)
		}
	}
}

func (this *ContractPayloadHandler) OnPayloadRejected(data dag.PayloadData) {

	log.I("[contract] on payload rejected. data length=", len(data))
	if data != nil {

		result := &ContractPayloadDagResult{ ResultType:ContractPayloadDagResultType_DagRejected, PayloadData:data }
		resultBytes, err := proto.Marshal(result)

		if err == nil {
			this.chanReceivedPayloads.Push(resultBytes)
		}
	}
}

func (this *ContractPayloadHandler) ProcessPayloadResult(data []byte) {

	log.I("[contract][processing payload] start.")
	result := &ContractPayloadDagResult{}
	err := proto.Unmarshal(data, result)
	if err != nil {
		log.W("[contract][processing payload] data is corrupted.")
	}


	if result.PayloadData == nil {

		log.W("[contract][process payload result] payload data is nil.")
		return
	}

	payload := &ContractPayload{}
	err = proto.Unmarshal(result.PayloadData, payload)

	log.I("[contract][processing payload] payload id=", hex.EncodeToString(payload.PayloadId), "result=", result.ResultType)

	if result.ResultType == ContractPayloadDagResultType_DagAccepted {

		if payload != nil && payload.Validate(this.contractStorage) {

			payload.Process(this.contractStorage)
			payload.PayloadStatus = ContractPayloadStatus_Accepted

		} else {
			payload.PayloadStatus = ContractPayloadStatus_Rejected
		}
	} else {
		payload.PayloadStatus = ContractPayloadStatus_Rejected_By_Dag
	}

	payload.ConfirmedTime, _ = ptypes.TimestampProto(time.Now())
	err = SavePayload(this.contractStorage, payload)
}