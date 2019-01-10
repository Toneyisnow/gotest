package dag

import (
	"../storage"
	"errors"
	"github.com/smartswarm/go/log"
	"strconv"
	"sync"
)

const (
	PendingPayloadBufferSize = 10

	IncomingVertexChannelCapacity = 10000
	SettledVertexChannelCapacity = 100
	SettledCandidateChannelCapacity = 100
	SettledQueenChannelCapacity = 100

	QueueCapacity = 10000
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

	// All queues defined
	queuePendingData *storage.RocksSequenceQueue

	// All channels defined
	chanIncomingVertex *storage.RocksChannel
	chanSettledVertex *storage.RocksChannel
	chanSettledCandidate *storage.RocksChannel
	chanSettledQueen *storage.RocksChannel

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

	dagStorage.queuePendingData = storage.NewRocksSequenceQueue(dagStorage.storage, "P", PendingPayloadBufferSize)

	// Loading the Channel data
	dagStorage.chanIncomingVertex = storage.NewRocksChannel(dagStorage.storage, "I", IncomingVertexChannelCapacity)
	dagStorage.chanIncomingVertex.Reload()

	dagStorage.chanSettledVertex = storage.NewRocksChannel(dagStorage.storage, "SV", SettledVertexChannelCapacity)
	dagStorage.chanSettledVertex.Reload()

	dagStorage.chanSettledCandidate = storage.NewRocksChannel(dagStorage.storage, "SC", SettledCandidateChannelCapacity)
	dagStorage.chanSettledCandidate.Reload()

	dagStorage.chanSettledQueen = storage.NewRocksChannel(dagStorage.storage, "SQ", SettledQueenChannelCapacity)
	dagStorage.chanSettledQueen.Reload()


	dagStorage.queueIncomingVertex = storage.NewRocksSequenceQueue(dagStorage.storage, "I", QueueCapacity)
	dagStorage.queueVertexDag = storage.NewRocksSequenceQueue(dagStorage.storage, "D", QueueCapacity)
	dagStorage.queueCandidate = storage.NewRocksSequenceQueue(dagStorage.storage, "C", QueueCapacity)
	dagStorage.queueQueen = storage.NewRocksSequenceQueue(dagStorage.storage, "Q", QueueCapacity)

	dagStorage.levelqueueUndecidedCandidate = storage.NewRocksLevelQueue(dagStorage.storage, "UC", QueueCapacity)
	dagStorage.queueUnconfirmedVertex = storage.NewRocksSequenceQueue(dagStorage.storage, "U", QueueCapacity)

	return dagStorage
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

