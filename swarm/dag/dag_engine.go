package dag

import (
	"../network"
	"../storage"
	"errors"
	"github.com/gogo/protobuf/proto"
	"github.com/smartswarm/go/log"
	"time"
)

type PayloadData []byte

type DagEngineStatus int
const (
	DagEngineStatus_Idle = 1
	DagEngineStatus_Started = 2
	DagEngineStatus_Connected = 3
)

type DagEngine struct {

	payloadHandler PayloadHandler

	// payloadDataQueueMutex    sync.Mutex
	// pendingPayloadDataQueues []PayloadData

	dagNodes    *DagNodes
	netTopology *network.NetTopology
	dagConfig   *DagConfig

	dagStorage *DagStorage

	incomingVertexDependency *storage.DependencyNotifier
	processVertexDependency *storage.DependencyNotifier

	// Stage 1: IncomingVertexChan: collect and save the vertexes that received from other nodes
	// incomingVertexChan *storage.RocksChannel

	// Stage 2: SettledVertexChan: the vertexes that saved in topology order waiting to be processed
	//settledVertexChan *storage.RocksChannel

	// Stage 3: FreshCandidateQueue: the candidates that waiting to be processed
	//freshCandidateQueue chan []byte

	// Stage 4: QueenQueue: the queue of the queens that could confirm the rest of the vertexes Accepted/Rejected
	//queenQueue chan []byte

	netProcessor *network.NetProcessor

	EngineStatus DagEngineStatus
}

func NewDagEngine(config *DagConfig, handler PayloadHandler) *DagEngine {

	engine := new(DagEngine)
	engine.dagConfig = config
	engine.BindHandler(handler)
	engine.Initialize()

	return engine
}

func (this *DagEngine) BindHandler(handler PayloadHandler) {

	this.payloadHandler = handler
}

func (this *DagEngine) Initialize() {

	this.dagNodes = &this.dagConfig.DagNodes
	this.netTopology = this.dagNodes.GetNetTopology()

	this.EngineStatus = DagEngineStatus_Idle

	// Load the previous cache data from Database: PendingPayloadData
	this.dagStorage = DagStorageGetInstance(this.dagConfig.StorageLocation)

	eventHandler := NewDagEventHandler(this)
	this.netProcessor = network.CreateProcessor(this.netTopology, eventHandler)

	this.incomingVertexDependency = storage.NewDependencyNotifier(this.dagStorage.chanIncomingVertex.Push)
	this.processVertexDependency = storage.NewDependencyNotifier(this.dagStorage.chanSettledVertex.Push)
}

func (this *DagEngine) Start() {

	log.I("[dag] begin DagEngine.Start")

	EnsureGenesisVertex(this.dagStorage, this.dagNodes.GetSelf())

	this.netProcessor.StartServer()
	this.EngineStatus = DagEngineStatus_Started

	for {
		this.RefreshConnection()
		if this.EngineStatus == DagEngineStatus_Connected {
			break
		}
		time.Sleep(3 * time.Second)
	}

	// Run every processing threads
	go this.dagStorage.chanIncomingVertex.Listen(this.OnIncomingVertex)
	go this.dagStorage.chanSettledVertex.Listen(this.OnSettledVertex)
	go this.dagStorage.chanSettledQueen.Listen(this.OnSettleQueen)

	// Keep making sure of the connection
	go func() {
		for {
			this.RefreshConnection()
			time.Sleep(3 * time.Second)
		}
	}()

	log.I("[dag] end DagEngine.Start")
}

func (this *DagEngine) Stop() {

}

func (this *DagEngine) RefreshConnection() {

	//log.I("[dag] Refreshing connection...")

	// Establish connection to all other nodes, and return success if connected to majority of nodes
	totalCount := 0
	successConnectionCount := 1		// Count itself as one of connections
	resultChans := make(chan bool)

	for _, node := range this.dagNodes.Peers {

		go func(device network.NetDevice, resultChan chan bool) {
			//log.I("[dag] Trying to connect to device", device.Id, " ...")
			result := <-this.netProcessor.ConnectToDeviceAsync(&device)
			if result {
				// log.I("Connecting to device ", device.Id, " succeed.")
				resultChans <- true
				return
			} else {
				// log.W("Connecting to device ", device.Id, " failed.")
				resultChans <- false
			}
		}(*node.Device, resultChans)
	}

	expectedConnectionCount := this.dagNodes.GetMajorityCount()
	for {
		result :=<-resultChans

		totalCount ++
		if result {
			successConnectionCount ++
		}

		if successConnectionCount >= expectedConnectionCount {

			if this.EngineStatus != DagEngineStatus_Connected {
				log.I("[dag] engine connected to dag. ConnectionCount=", successConnectionCount, "Expected=", expectedConnectionCount)
			}
			this.EngineStatus = DagEngineStatus_Connected
			break
		}

		if totalCount >= len(this.dagNodes.Peers) {
			// All Peers returned, but not enough success connection
			if this.EngineStatus == DagEngineStatus_Connected {
				log.I("[dag] engine disconnected from dag, connection is not enough. ConnectionCount=", successConnectionCount, "Expected=", expectedConnectionCount)
			}
			this.EngineStatus = DagEngineStatus_Started
			break
		}
	}
}

//
func (this *DagEngine) SubmitPayload(data PayloadData) (err error) {

	log.I("[dag] begin submit payload.")

	if this.EngineStatus != DagEngineStatus_Connected {
		log.W("[dag] cannot submit payload since the dagEngine is not connected to dag.")
		return errors.New("cannot submit payload since dagEngine is not connected to dag")
	}

	createdVertex, err := CreateSelfDataVertex(dagStorage, this.dagNodes.GetSelf(), data)
	if createdVertex != nil {

		// Handle the event
		if this.payloadHandler != nil {
			for _, payloadData := range createdVertex.GetContent().Data {
				this.payloadHandler.OnPayloadSubmitted(payloadData)
			}
		}

		this.composeVertexEvent(createdVertex)
	}

	log.I("[dag] end submit payload.")
	return nil
}

func (this *DagEngine) composeVertexEvent(mainVertex *DagVertex) {

	log.I("[dag] begin compose vertex event.")

	if mainVertex == nil {
		log.W("[dag] mainVertex is nil, cannot compose vertex event.")
		return
	}

	// Send the Vertex to 0-2 nodes
	peerNodes := SelectPeerNodeToSendVertex(this.dagStorage, mainVertex, this.dagNodes)
	if peerNodes != nil && len(peerNodes) != 0 {

		for _, peerNode := range peerNodes {

			// Compose Vertex event and send
			relatedVertexes, _ := FindPossibleUnknownVertexesForNode(this.dagStorage, peerNode)
			event, _ := NewVertexEvent(mainVertex, relatedVertexes)
			data, _ := proto.Marshal(event)

			resultChan := this.netProcessor.SendEventToDeviceAsync(peerNode.Device, data)
			result := <-resultChan

			if (result.Err != nil) {
				log.I2("[dag] send event finished. result: eventId=[%d], err=[%s]", result.EventId, result.Err.Error())
			} else {
				log.I2("[dag] send event succeeded. result: eventId=[%d]", result.EventId)
			}
		}
	}

	log.I("[dag] end composeVertexEvent.")
}

// Threads from Workers: create vertex with this as peer parent if possible
func (this *DagEngine) PushIncomingMainVertex(vertex *DagVertex) (err error) {

	log.I("[dag] begin push incoming vertex.")

	if this.EngineStatus != DagEngineStatus_Connected {
		log.W("[dag] cannot submit payload since the dagEngine is not connected to dag.")
		return errors.New("cannot submit payload since dagEngine is not connected to dag")
	}

	if vertex == nil {
		log.W("[dag] cannot push nil main vertex.")
		return errors.New("cannot push nil main vertex")
	}

	createdVertex, err := CreateTwoParentsVertex(dagStorage, this.dagNodes.GetSelf(), vertex)
	if createdVertex != nil {

		// Handle the event
		if this.payloadHandler != nil {
			for _, payloadData := range createdVertex.GetContent().Data {
				this.payloadHandler.OnPayloadSubmitted(payloadData)
			}
		}

		this.composeVertexEvent(createdVertex)
	}

	log.I("[dag] end push incoming vertex.")
	return nil
}

// Thread 1: Validate the incoming vertexes and build the dag
func (this *DagEngine) OnIncomingVertex(data []byte) {

	log.I("[dag] begin OnIncomingVertex")

	vertex := &DagVertex{}
	proto.Unmarshal(data, vertex)

	decision, missingParentVertex := ProcessIncomingVertex(this.dagStorage, this.dagNodes, vertex)

	switch decision {
		case ProcessResult_Yes:
			// Push to next channel to process
			this.dagStorage.chanSettledVertex.Push(vertex.Hash)

			// Also notify all the dependency vertexes to handle
			this.incomingVertexDependency.Notify(vertex.Hash)
			break
		case ProcessResult_No:
			// Discard this vertex if it's invalid
			break
		case ProcessResult_Undecided:
			// If it's undecided, put it back into the incoming queue
			this.incomingVertexDependency.SetDependency(missingParentVertex, data)
			break
	}

	log.I("[dag] end OnIncomingVertex")
}

// Thread 2: Mark vertex levels, and decide the candidates
func (this *DagEngine) OnSettledVertex(hash []byte) {

	log.I("[dag] begin OnSettledVertex")

	result, missingParentHash := ProcessVertexAndDecideCandidate(this.dagStorage, this.dagNodes, hash)
	log.I("ProcessVertexAndDecideCandidate: result=", result)

	switch result {
		case ProcessResult_Yes:
			// Push to next channel to process
			this.dagStorage.chanSettledCandidate.Push(hash)

			// Also notify all the dependency vertexes to handle
			this.processVertexDependency.Notify(hash)
			break
		case ProcessResult_No:
			break
		case ProcessResult_Undecided:
			this.processVertexDependency.SetDependency(missingParentHash, hash)
			break
	}
	log.I("[dag] end OnSettledVertex")
}

// Thread 3: Candidates vote for Queen and Decide Queen
func (this *DagEngine) OnSettledCandidate(candidateHash []byte) {

	log.I("[dag] begin OnSettledCandidate")

	// The Collect vote method will decide new queens, and write into the chanSettledQueen
	ProcessCandidateVote(this.dagStorage, this.dagNodes, candidateHash, func(queenHash []byte) {
		this.dagStorage.chanSettledQueen.Push(queenHash)
	})

	log.I("[dag] end OnSettledCandidate")
}

// Thread 4: Queen to decide accept/reject vertex
func (this *DagEngine) OnSettleQueen(queenHash []byte) {

	log.I("[dag] begin OnSettleQueen")

	completed := ProcessQueenDecision(this.dagStorage, this.dagNodes, queenHash, func (vertexHash []byte, result VertexConfirmResult) {

		// On Vertex Accepted
		vertex := GetVertex(this.dagStorage, vertexHash)
		if vertex == nil || vertex.GetContent() == nil || vertex.GetContent().GetData() == nil {
			return
		}

		dataList := vertex.GetContent().GetData()
		for _, data := range dataList {
			this.payloadHandler.OnPayloadAccepted(data)
		}
	})

	if completed == ProcessResult_Yes {

	} else {

		// If something wrong happened in processing, put the queen back to the queue and try again
		this.dagStorage.chanSettledQueen.Push(queenHash)
	}

	log.I("[dag] end OnSettleQueen")
}

