package contract

import (
	"../storage"
	"github.com/gogo/protobuf/proto"
	"sync"
)

const (

	QueueCapacity = 10000
)

type ContractStorage struct {

	osFilelocation string

	storage *storage.RocksStorage

	// All tables defined
	tablePayload *storage.RocksTable
	tableWallet *storage.RocksTable
	tableAsset *storage.RocksTable

	// All Channels defined
	chanReceivedPayload *storage.RocksChannel

}

var contractStorage *ContractStorage
var contractStorageMutex sync.Mutex

func ContractStorageGetInstance(storageLocation string) *ContractStorage{

	contractStorageMutex.Lock()
	if contractStorage == nil {
		contractStorage = NewContractStorage(storageLocation)
	}
	contractStorageMutex.Unlock()

	return contractStorage
}

func NewContractStorage(storageLocation string) *ContractStorage {

	contractStorage := new(ContractStorage)

	contractStorage.osFilelocation = storageLocation
	contractStorage.storage = storage.ComposeRocksDBInstance(storageLocation + "swarmcontract")

	// ------ Initialize the table data ------

	// Payload Table: key:[id] value:[payload_bytes]
	contractStorage.tablePayload = storage.NewRocksTable(contractStorage.storage, "P")

	// Wallet Table: key:[wallet_address] value:[wallet_hash]
	contractStorage.tableWallet = storage.NewRocksTable(contractStorage.storage, "W")

	// Asset Table: key:[asset_type+asset_id] value:[asset_hash]
	contractStorage.tableAsset = storage.NewRocksTable(contractStorage.storage, "A")

	// ------ Initialize the channel data ------

	contractStorage.chanReceivedPayload = storage.NewRocksChannel(contractStorage.storage, "RP", QueueCapacity)

	return contractStorage
}

func SavePayload(contractStorage *ContractStorage, payload *ContractPayload) error {

	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return err
	}

	err = contractStorage.tablePayload.InsertOrUpdate(payload.PayloadId, payloadBytes)
	return err

}

func GetPayload(contractStorage *ContractStorage, id PayloadId) *ContractPayload {

	payloadBytes := contractStorage.tablePayload.Get(id)
	if payloadBytes == nil {
		return nil
	}

	payload := &ContractPayload{}
	err := proto.Unmarshal(payloadBytes, payload)
	if err != nil {
		return nil
	}

	return payload
}