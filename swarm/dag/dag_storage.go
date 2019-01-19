package dag

import (
	"../storage"
	"github.com/gogo/protobuf/proto"
	"sync"
)

const (
	PendingPayloadBufferSize = 3

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
	tableGenesisVertex *storage.RocksTable

	tableVertexLink *storage.RocksTable
	tableVertexStatus *storage.RocksTable
	tableVertexConnection *storage.RocksTable
	tableCandidateDecision *storage.RocksTable

	tableNodeSyncTimestamp *storage.RocksTable

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
	if dagStorage == nil {
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
	dagStorage.tableCandidate = storage.NewRocksTable(dagStorage.storage, "VC")

	// Last Vertex Table: key:[nodeId] value:[vertex_hash]
	dagStorage.tableNodeLatestVertex = storage.NewRocksTable(dagStorage.storage, "NV")

	// Genesis Vertex Table: key:[nodeId] value:[vertex_hash]
	dagStorage.tableGenesisVertex = storage.NewRocksTable(dagStorage.storage, "GV")

	// Vertex Parent Table: key:[vertex_hash] value:[self_parent_hash+peer_parent_hash]
	dagStorage.tableVertexLink = storage.NewRocksTable(dagStorage.storage, "VD")

	// Vertex Status Table: key:[vertex_hash] value:[level+isCandidate+isQueen+status]
	dagStorage.tableVertexStatus = storage.NewRocksTable(dagStorage.storage, "VS")

	// Vertex Connection Table: key:[vertex_hash+vertex_hash] value:[nodeId1, nodeId2, ...]
	dagStorage.tableVertexConnection = storage.NewRocksTable(dagStorage.storage, "VR")

	// Candidate Decision Table: key:[vertex_hash+vertex_hash] value:[Yes, No, DecideYes, DecideNo]
	dagStorage.tableCandidateDecision = storage.NewRocksTable(dagStorage.storage, "CV")

	// The last synced timestamp for a given node
	dagStorage.tableNodeSyncTimestamp = storage.NewRocksTable(dagStorage.storage, "NT")


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


func GetVertex(dagStorage *DagStorage, vertexHash []byte) *DagVertex {

	vertexByte := dagStorage.tableVertex.Get(vertexHash)
	if vertexByte == nil {
		return nil
	}

	vertex := &DagVertex{}
	err := proto.Unmarshal(vertexByte, vertex)
	if err != nil {
		return nil
	}

	return vertex
}

func SaveVertex(dagStorage *DagStorage, vertex *DagVertex) (err error) {

	vertexBytes, err := proto.Marshal(vertex)
	if err != nil {
		return err
	}

	err = dagStorage.tableVertex.InsertOrUpdate(vertex.GetHash(), vertexBytes)
	return err
}

func GetVertexStatus(dagStorage *DagStorage, vertexHash []byte) *DagVertexStatus {

	statusByte := dagStorage.tableVertexStatus.Get(vertexHash)
	if statusByte == nil {
		return nil
	}

	status := &DagVertexStatus{}
	err := proto.Unmarshal(statusByte, status)
	if err != nil {
		return nil
	}

	return status
}

func SetVertexStatus(dagStorage *DagStorage, vertexHash []byte, status *DagVertexStatus) {

	statusByte, _ := proto.Marshal(status)
	err := dagStorage.tableVertexStatus.InsertOrUpdate(vertexHash, statusByte)
	if err != nil {

	}
}

func SetLatestNodeVertex(dagStorage *DagStorage, nodeId uint64, vertexHash []byte) {

	key := storage.ConvertUint64ToBytes(nodeId)
	err := dagStorage.tableNodeLatestVertex.InsertOrUpdate(key, vertexHash)
	if err != nil {

	}
}

func GetLatestNodeVertex(dagStorage *DagStorage, nodeId uint64, hashOnly bool) (hash []byte, vertex *DagVertex) {

	key := storage.ConvertUint64ToBytes(nodeId)
	resultByte := dagStorage.tableNodeLatestVertex.Get(key)
	if resultByte == nil {
		return nil, nil
	}

	if hashOnly {
		return resultByte, nil
	}

	vertexByte := dagStorage.tableVertex.Get(resultByte)

	result := &DagVertex{}
	err := proto.Unmarshal(vertexByte, result)
	if err != nil {
		return nil, nil
	}

	return resultByte, result
}

func GetVertexLink(dagStorage *DagStorage, vertexHash []byte) *DagVertexLink {

	linkByte := dagStorage.tableVertexStatus.Get(vertexHash)
	if linkByte == nil {
		return nil
	}

	link := &DagVertexLink{}
	err := proto.Unmarshal(linkByte, link)
	if err != nil {
		return nil
	}

	return link
}

func GetCandidateForNode(dagStorage *DagStorage, nodeId uint64, level uint32, hashOnly bool) (hash []byte, vertex *DagVertex) {

	key := append(storage.ConvertUint64ToBytes(nodeId), storage.ConvertUint32ToBytes(level)...)
	resultByte := dagStorage.tableVertexStatus.Get(key)
	if resultByte == nil {
		return nil, nil
	}

	if hashOnly {
		return resultByte, nil
	}

	vertexByte := dagStorage.tableVertex.Get(resultByte)

	result := &DagVertex{}
	err := proto.Unmarshal(vertexByte, result)
	if err != nil {
		return nil, nil
	}

	return resultByte, result
}


func GetGenesisVertex(dagStorage *DagStorage, nodeId uint64, hashOnly bool) (hash []byte, vertex *DagVertex) {

	key := storage.ConvertUint64ToBytes(nodeId)
	resultByte := dagStorage.tableGenesisVertex.Get(key)
	if resultByte == nil {
		return nil, nil
	}

	if hashOnly {
		return resultByte, nil
	}

	vertexByte := dagStorage.tableVertex.Get(resultByte)

	result := &DagVertex{}
	err := proto.Unmarshal(vertexByte, result)
	if err != nil {
		return nil, nil
	}

	return resultByte, result
}

func GetVertexConnection(dagStorage *DagStorage, vertexHash []byte, targetVertexHash []byte) *DagVertexConnection {

	key := append(vertexHash, targetVertexHash...)
	resultByte := dagStorage.tableVertexStatus.Get(key)
	if resultByte == nil {
		return nil
	}

	result := &DagVertexConnection{}
	err := proto.Unmarshal(resultByte, result)
	if err != nil {
		return nil
	}

	return result
}


func GetCandidateDecision(dagStorage *DagStorage, vertexHash []byte, targetVertexHash []byte) CandidateDecision {

	key := append(vertexHash, targetVertexHash...)
	resultByte := dagStorage.tableCandidateDecision.Get(key)
	if resultByte == nil {
		return CandidateDecision_Unknown
	}

	result := storage.ConvertBytesToUint32(resultByte)
	return CandidateDecision(result)
}


func SetCandidateDecision(dagStorage *DagStorage, vertexHash []byte, targetVertexHash []byte, decision CandidateDecision) {

	key := append(vertexHash, targetVertexHash...)
	value := storage.ConvertUint32ToBytes(uint32(decision))

	dagStorage.tableCandidateDecision.InsertOrUpdate(key, value)
}
