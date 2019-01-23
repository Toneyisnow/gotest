package storage

import (
	"sync"
)

type RocksTable struct {

	RocksContainer

	tableId string
	tableUpdateMutex sync.Mutex
}

func NewRocksTable(st *RocksStorage, tableId string) *RocksTable {

	queue := new(RocksTable)

	queue.storage = st
	queue.containerId = tableId
	queue.containerType = RocksContainerType_Table

	return queue
}

func (this *RocksTable) InsertOrUpdate(key []byte, value []byte) (err error) {

	this.tableUpdateMutex.Lock()
	defer this.tableUpdateMutex.Unlock()

	this.storage.Put(this.GenerateKey(key), value)

	return nil
}

func (this *RocksTable) Get(key []byte) []byte {

	value, err := this.storage.Get(this.GenerateKey(key))
	if err != nil {
		return nil
	}

	return value
}

func (this *RocksTable) Exists(key []byte) bool {

	value := this.Get(key)
	return value != nil
}

func (this *RocksTable) Delete(key []byte) error {

	this.tableUpdateMutex.Lock()
	defer this.tableUpdateMutex.Unlock()

	err := this.storage.Del(this.GenerateKey(key))
	return err
}
