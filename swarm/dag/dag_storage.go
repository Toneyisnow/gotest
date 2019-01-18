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

	osFilelocation string

	storage *storage.RocksStorage

	// All tables defined
	tableVertex *storage.RocksTable
	tableCandidate *storage.RocksTable
	tableNodeLatestVertex *storage.RocksTable

	tableVertexLink *storage.RocksTable
	tableVertexStatus *storage.RocksTable
	tableVertexConnection *storage.RocksTable
	tableCandidateDecision *storage.RocksTable

	// All queues defined
	queuePendingData            *storage.RocksSequenceQueue

	// All levelQueues defined
	levelQueueUnconfirmedVertex   *storage.RocksLevelQueue
	levelQueueUndecidedCandidate  *storage.RocksLevelQueue

	// All channels defined
	chanIncomingVertex *storage.RocksChannel
	chanSettledVertex *storage.RocksChannel
	chanSettledCandidate *storage.RocksChannel
	chanSettledQueen *storage.RocksChannel

	/*
	// Queue: Incoming Vertex
	queueIncomingVertex *storage.RocksSequenceQueue
	queueVertexDag *storage.RocksSequenceQueue
	queueCandidate *storage.RocksSequenceQueue
	queueQueen *storage.RocksSequenceQueue
	*/

}

var dagStorage *DagStorage
var dagStorageMutex sync.Mutex

func DagStorageGetInstance(storageLocation string) *DagStorage{

	dagStorageMutex.Lock()
	if (dagStorage == nil) {
		dagStorage = NewDagStorage(storageLocation)
	}
	dagStorageMutex.Unlock()

	return dagStorage
}

func NewDagStorage(storageLocation string) *DagStorage {

	dagStorage := new(DagStorage)

	dagStorage.osFilelocation = storageLocation
	dagStorage.storage = storage.ComposeRocksDBInstance(storageLocation + "swarmdag")

	// ------ Initialize the table data ------

	// Vertex Table: key:[vertex_hash] value:[vertex_bytes]
	dagStorage.tableVertex = storage.NewRocksTable(dagStorage.storage, "V")

	// Candidate Table: key:[nodeId+level] value:[vertex_hash]
	dagStorage.tableCandidate = storage.NewRocksTable(dagStorage.storage, "C")

	// Last Vertex Table: key:[nodeId] value:[vertex_hash]
	dagStorage.tableNodeLatestVertex = storage.NewRocksTable(dagStorage.storage, "NV")

	// Vertex Parent Table: key:[vertex_hash] value:[self_parent_hash+peer_parent_hash]
	dagStorage.tableVertexLink = storage.NewRocksTable(dagStorage.storage, "VD")

	// Vertex Status Table: key:[vertex_hash] value:[level+isCandidate+isQueen+status]
	dagStorage.tableVertexStatus = storage.NewRocksTable(dagStorage.storage, "VS")

	// Vertex Connection Table: key:[vertex_hash+vertex_hash] value:[nodeId1, nodeId2, ...]
	dagStorage.tableVertexConnection = storage.NewRocksTable(dagStorage.storage, "VR")

	// Candidate Decision Table: key:[vertex_hash+vertex_hash] value:[Yes, No, DecideYes, DecideNo]
	dagStorage.tableCandidateDecision = storage.NewRocksTable(dagStorage.storage, "CV")


	// ------ Initialize the queue data ------
	dagStorage.queuePendingData = storage.NewRocksSequenceQueue(dagStorage.storage, "P", PendingPayloadBufferSize)


	// ------ Initialize the levelQueue data------
	dagStorage.levelQueueUnconfirmedVertex = storage.NewRocksLevelQueue(dagStorage.storage, "UV", QueueCapacity)
	dagStorage.levelQueueUndecidedCandidate = storage.NewRocksLevelQueue(dagStorage.storage, "UC", QueueCapacity)


	// ------  Loading the Channel data ------
	dagStorage.chanIncomingVertex = storage.NewRocksChannel(dagStorage.storage, "I", IncomingVertexChannelCapacity)
	dagStorage.chanIncomingVertex.Reload()

	dagStorage.chanSettledVertex = storage.NewRocksChannel(dagStorage.storage, "SV", SettledVertexChannelCapacity)
	dagStorage.chanSettledVertex.Reload()

	dagStorage.chanSettledCandidate = storage.NewRocksChannel(dagStorage.storage, "SC", SettledCandidateChannelCapacity)
	dagStorage.chanSettledCandidate.Reload()

	dagStorage.chanSettledQueen = storage.NewRocksChannel(dagStorage.storage, "SQ", SettledQueenChannelCapacity)
	dagStorage.chanSettledQueen.Reload()

	return dagStorage
}

func (this *DagStorage) GetLastVertexOnNode(node *DagNode, hashOnly bool) (hash []byte, vertex *DagVertex, err error) {

	if (node == nil) {
		log.W("GetLastVertexOnNode failed: node is nil.")
		return nil, nil, errors.New("GetLastVertexOnNode failed: node is nil.")
	}

	vertexHash, err := this.storage.Get([]byte("L:" + strconv.FormatUint(node.NodeId, 16)))

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

