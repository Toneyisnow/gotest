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

	log.I("[dag] begin dag engine start...")


	log.I("[dag] setting up the channel callbacks.")
	// Run every processing threads
	go this.dagStorage.chanIncomingVertex.Listen(this.OnIncomingVertex)
	go this.dagStorage.chanSettledVertex.Listen(this.OnSettledVertex)
	go this.dagStorage.chanSettledQueen.Listen(this.OnSettleQueen)

	this.netProcessor.StartServer()
	this.EngineStatus = DagEngineStatus_Started

	log.I("[dag] trying to connect to dag...")
	for {
		this.RefreshConnection()
		if this.EngineStatus == DagEngineStatus_Connected {
			break
		}
		time.Sleep(3 * time.Second)
	}

	isNewCreatedGenesis, genesisVertexHash := EnsureGenesisVertex(this.dagStorage, this.dagNodes.GetSelf())
	if isNewCreatedGenesis {
		incomingVertex := &DagVertexIncoming{ Hash:genesisVertexHash, IsMain:false }
		dagStorage.chanIncomingVertex.PushProto(incomingVertex)
	}

	// Keep making sure of the connection
	go func() {
		for {
			this.RefreshConnection()
			time.Sleep(3 * time.Second)
		}
	}()

	log.I("[dag] dag engine started.")
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

	log.I("[dag] submit payload...")

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

		// Should process the created vertex as incoming vertex
		this.dagStorage.chanIncomingVertex.PushProto(createdVertex)

		this.composeVertexEvent(createdVertex)
	}

	return nil
}

func (this *DagEngine) composeVertexEvent(mainVertex *DagVertex) {

	log.I("[dag] begin compose vertex event...")

	if mainVertex == nil {
		log.W("[dag] mainVertex is nil, cannot compose vertex event.")
		return
	}

	// Send the Vertex to 0-2 nodes
	peerNodes := SelectPeerNodeToSendVertex(this.dagStorage, mainVertex, this.dagNodes)
	log.I("[dag] decide to send to target nodes (count=", len(peerNodes), "):, peerNodes:", peerNodes)

	if peerNodes != nil && len(peerNodes) != 0 {

		for _, peerNode := range peerNodes {

			// Compose Vertex event and send
			relatedVertexes, _ := FindPossibleUnknownVertexesForNode(this.dagStorage, this.dagNodes.GetSelf(), peerNode)
			log.I("[dag] find possible unknown vertex for node", peerNode.NodeId, ", totally", len(relatedVertexes), "related vertexes found.")
			event, _ := NewVertexEvent(mainVertex, relatedVertexes)
			data, _ := proto.Marshal(event)

			resultChan := this.netProcessor.SendEventToDeviceAsync(peerNode.Device, data)
			result := <-resultChan

			if (result.Err != nil) {
				log.I2("[dag] send event finished. result: eventId=[",result.EventId,"], err=",  result.Err.Error())
				FlagKnownVertexForNode(this.dagStorage, peerNode, append(relatedVertexes, mainVertex))

			} else {
				log.I2("[dag] send event succeeded. result: eventId=[",result.EventId, "]")
			}
		}
	}
}

// Threads from Workers: create vertex with this as peer parent if possible
func (this *DagEngine) HandleIncomingMainVertex(vertexHash []byte) (err error) {

	log.I("[dag] handling incoming main vertex.")

	if this.EngineStatus != DagEngineStatus_Connected {
		log.W("[dag] cannot submit payload since the dagEngine is not connected to dag.")
		return errors.New("cannot submit payload since dagEngine is not connected to dag")
	}

	if vertexHash == nil {
		log.W("[dag] cannot push nil main vertex.")
		return errors.New("cannot push nil main vertex")
	}

	createdVertex, err := CreateTwoParentsVertex(dagStorage, this.dagNodes.GetSelf(), vertexHash)
	if createdVertex != nil {

		// Handle the event
		if this.payloadHandler != nil {
			for _, payloadData := range createdVertex.GetContent().Data {
				this.payloadHandler.OnPayloadSubmitted(payloadData)
			}
		}

		this.composeVertexEvent(createdVertex)
	}

	/// log.I("[dag] end push incoming vertex.")
	return nil
}

// Thread 1: Validate the incoming vertexes and build the dag
func (this *DagEngine) OnIncomingVertex(data []byte) {

	log.I("[dag] handling incoming vertex...")

	incomingVertex := &DagVertexIncoming{}
	err := proto.Unmarshal(data, incomingVertex)
	if err != nil || incomingVertex.Hash == nil {
		log.W("[dag] unmarshal incoming vertex failed, skip it")
		return
	}

	decision, missingParentVertex := ProcessIncomingVertex(this.dagStorage, this.dagNodes, incomingVertex.Hash)
	log.I("[dag] processing incoming vertex",GetShortenedHash(incomingVertex.Hash)," result:", decision)

	switch decision {
		case ProcessResult_Yes:
			// Push to next channel to process
			this.dagStorage.chanSettledVertex.Push(incomingVertex.Hash)

			// Also notify all the dependency vertexes to handle
			this.incomingVertexDependency.Notify(incomingVertex.Hash)

			if incomingVertex.IsMain {
				err = this.HandleIncomingMainVertex(incomingVertex.Hash)
			}
			break
		case ProcessResult_No:
			// Discard this vertex if it's invalid
			break
		case ProcessResult_Undecided:
			// If it's undecided, put it back into the incoming queue
			this.incomingVertexDependency.SetDependency(missingParentVertex, data)
			break
	}

	/// log.I("[dag] end OnIncomingVertex")
}

// Thread 2: Mark vertex levels, and decide the candidates
func (this *DagEngine) OnSettledVertex(hash []byte) {

	log.I("[dag] on settled vertex...")

	result, missingParentHash := ProcessVertexAndDecideCandidate(this.dagStorage, this.dagNodes, hash)
	log.I("[dag] process vertex decide candidate result:", result)

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
			if missingParentHash != nil {
				this.processVertexDependency.SetDependency(missingParentHash, hash)
			}
			break
	}
}

// Thread 3: Candidates vote for Queen and Decide Queen
func (this *DagEngine) OnSettledCandidate(candidateHash []byte) {

	log.I("[dag] on settled candidate...")

	// The Collect vote method will decide new queens, and write into the chanSettledQueen
	ProcessCandidateVote(this.dagStorage, this.dagNodes, candidateHash, func(queenHash []byte) {
		this.dagStorage.chanSettledQueen.Push(queenHash)
	})
}

// Thread 4: Queen to decide accept/reject vertex
func (this *DagEngine) OnSettleQueen(queenHash []byte) {

	log.I("[dag] on settle queen...")

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
}

