package dag

import (
	"../storage"
	"math/rand"
	"strconv"
)

type DagStorage struct {

	_storage *storage.RocksStorage
}

func ComposeDagStorageInstance() *DagStorage {

	dagStorage := new(DagStorage)

	dagStorage._storage = storage.ComposeRocksDBInstance("swarmdag")

	return dagStorage
}

func (this *DagStorage) PutPendingPayloadData(data PayloadData) {

	key := "P:" + strconv.Itoa(rand.Int())
	this._storage.PutSeek([]byte(key), data)
}

func (this *DagStorage) GetPendingPayloadData() []PayloadData {

	result := make([]PayloadData, 0)

	this._storage.Seek([]byte("P:"), func(v []byte) {
		result = append(result, v)
	})

	return result
}
