package contract

import (
	"github.com/golang/protobuf/ptypes"
	"math/rand"
	"strconv"
	"time"
)

func NewDumbPayload() (result *ContractPayload) {

	payload := new(ContractPayload)
	payload.PayloadId = GeneratePayloadId()
	payload.CreatedTime, _ = ptypes.TimestampProto(time.Now())
	payload.PayloadStatus = ContractPayloadStatus_Initialized
	payload.PayloadType = ContractPayloadType_Dumb

	dumb := new(DumbPayload)
	dumb.RandomString = strconv.FormatUint(rand.Uint64(), 10)

	payload.Data = &ContractPayload_DumbPayload{DumbPayload: dumb}

	return payload;
}

func (this *DumbPayload) Validate(storage *ContractStorage) bool {

	return true
}

func (this *DumbPayload) Process(storage *ContractStorage) error {

	// Nothing to do

	return nil
}
