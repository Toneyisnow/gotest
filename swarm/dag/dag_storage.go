package dag

import (
	"../storage"
	"errors"
	"github.com/smartswarm/go/log"
	"math/rand"
	"strconv"
	"sync"
)

type DagStorage struct {

	_storage *storage.RocksStorage
}

var _dagStorage *DagStorage
var _dagStorageMutex sync.Mutex

func DagStorageGetInstance() *DagStorage{

	_dagStorageMutex.Lock()
	if (_dagStorage == nil) {
		_dagStorage = ComposeDagStorageInstance()
	}
	_dagStorageMutex.Unlock()

	return _dagStorage
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

func (this *DagStorage) GetLastVertexOnNode(node *DagNode, hashOnly bool) (hash string, vertex *DagVertex, err error) {

	if (node == nil) {
		log.W("GetLastVertexOnNode failed: node is nil.")
		return "", nil, errors.New("GetLastVertexOnNode failed: node is nil.")
	}

	vertexHash, err := this._storage.Get([]byte("L:" + strconv.FormatInt(node.NodeId, 16)))

	if err != nil {
		log.W("Got error while GetLastVertexOnNode: " + err.Error())
		return "", nil, err
	}

	if vertexHash == nil {
		log.W("Cannot find vertex hash in GetLastVertexOnNode.")
		return "", nil, errors.New("Cannot find vertex hash in GetLastVertexOnNode.")
	}

	if hashOnly {
		return string(vertexHash), nil, nil
	}

	vertex = new(DagVertex)
	this._storage.LoadProto("V:" + string(vertexHash), vertex)
	if vertex == nil {
		log.W("Cannot find vertex in GetLastVertexOnNode.")
		return "", nil, errors.New("Cannot find vertex in GetLastVertexOnNode.")
	}

	return string(vertexHash), vertex, nil
}

