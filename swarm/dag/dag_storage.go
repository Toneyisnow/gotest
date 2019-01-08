package dag

import (
	"../storage"
	"errors"
	"github.com/smartswarm/go/log"
	"strconv"
	"sync"
)

type DagStorage struct {

	storage *storage.RocksStorage

	// All tables defined
	tableVertex *storage.RocksTable
	tableVertexDag *storage.RocksTable
	tableVertexStatus *storage.RocksTable
	tableCandidate *storage.RocksTable
	tableNodeLatestVertex *storage.RocksTable
	tableCandidateVote *storage.RocksTable

	queuePendingData *storage.RocksSequenceQueue

	// Queue: Incoming Vertex
	queueIncomingVertex *storage.RocksSequenceQueue
	queueVertexDag *storage.RocksSequenceQueue
	queueCandidate *storage.RocksSequenceQueue
	queueQueen *storage.RocksSequenceQueue

	levelqueueUndecidedCandidate *storage.RocksLevelQueue
	queueUnconfirmedVertex *storage.RocksSequenceQueue
	
}

var dagStorage *DagStorage
var dagStorageMutex sync.Mutex

func DagStorageGetInstance() *DagStorage{

	dagStorageMutex.Lock()
	if (dagStorage == nil) {
		dagStorage = NewDagStorage()
	}
	dagStorageMutex.Unlock()

	return dagStorage
}

func NewDagStorage() *DagStorage {

	dagStorage := new(DagStorage)

	dagStorage.storage = storage.ComposeRocksDBInstance("swarmdag")

	dagStorage.queuePendingData = storage.NewRocksSequenceQueue(dagStorage.storage, "P")
	dagStorage.queueIncomingVertex = storage.NewRocksSequenceQueue(dagStorage.storage, "I")
	dagStorage.queueVertexDag = storage.NewRocksSequenceQueue(dagStorage.storage, "D")
	dagStorage.queueCandidate = storage.NewRocksSequenceQueue(dagStorage.storage, "C")
	dagStorage.queueQueen = storage.NewRocksSequenceQueue(dagStorage.storage, "Q")

	dagStorage.levelqueueUndecidedCandidate = storage.NewRocksLevelQueue(dagStorage.storage, "UC")
	dagStorage.queueUnconfirmedVertex = storage.NewRocksSequenceQueue(dagStorage.storage, "U")

	return dagStorage
}

func (this *DagStorage) PutPendingPayloadData(data PayloadData) {

	this.queuePendingData.Push(data)
}

func (this *DagStorage) GetPendingPayloadData() []PayloadData {

	this.queuePendingData.Pop()

	result := make([]PayloadData, 0)

	this.storage.SeekAll([]byte("P:"), func(v []byte) {
		result = append(result, v)
	})

	return result
}

func (this *DagStorage) GetLastVertexOnNode(node *DagNode, hashOnly bool) (hash []byte, vertex *DagVertex, err error) {

	if (node == nil) {
		log.W("GetLastVertexOnNode failed: node is nil.")
		return nil, nil, errors.New("GetLastVertexOnNode failed: node is nil.")
	}

	vertexHash, err := this.storage.Get([]byte("L:" + strconv.FormatInt(node.NodeId, 16)))

	if err != nil {
		log.W("Got error while GetLastVertexOnNode: " + err.Error())
		return nil, nil, err
	}

	if vertexHash == nil {
		log.W("Cannot find vertex hash in GetLastVertexOnNode.")
		return nil, nil, errors.New("Cannot find vertex hash in GetLastVertexOnNode.")
	}

	if hashOnly {
		return vertexHash, nil, nil
	}

	vertex = new(DagVertex)
	this.storage.LoadProto("V:" + string(vertexHash), vertex)
	if vertex == nil {
		log.W("Cannot find vertex in GetLastVertexOnNode.")
		return nil, nil, errors.New("Cannot find vertex in GetLastVertexOnNode.")
	}

	return vertexHash, vertex, nil
}

