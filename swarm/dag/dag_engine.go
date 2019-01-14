package dag

import (
	"../network"
	"../storage"
	"github.com/gogo/protobuf/proto"
	"github.com/smartswarm/go/log"
	"os"
	"strconv"
)

type PayloadData []byte


type DagEngine struct {

	payloadHandler PayloadHandler

	// payloadDataQueueMutex    sync.Mutex
	// pendingPayloadDataQueues []PayloadData

	dagNodes    *DagNodes
	netTopology *network.NetTopology

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
}

func ComposeDagEngine(handler PayloadHandler) *DagEngine {

	engine := new(DagEngine)
	engine.BindHandler(handler)
	engine.Initialize()

	return engine
}

func (this *DagEngine) BindHandler(handler PayloadHandler) {

	this.payloadHandler = handler
}

func (this *DagEngine) Initialize() {

	// this.pendingPayloadDataQueues = make([]PayloadData, 0)

	this.dagNodes = LoadDagNodesFromFile("node-topology.json")
	this.netTopology = this.dagNodes.GetNetTopology()

	if (len(os.Args) > 1) {
		serverPort, _ := strconv.Atoi(os.Args[1])
		this.netTopology.Self().Port = int32(serverPort)
	}

	// Create Channels
	// this.incomingVertexChan = storage.NewRocksChannel(this.dagStorage.storage, "")

	//this.settledVertexChan = new(storage.RocksChannel)

	//this.freshCandidateQueue = make(chan []byte)
	//this.queenQueue = make(chan []byte)


	this.dagStorage = DagStorageGetInstance()

	eventHandler := NewDagEventHandler(this)
	this.netProcessor = network.CreateProcessor(this.netTopology, eventHandler)

	this.incomingVertexDependency = storage.NewDependencyNotifier(this.dagStorage.chanIncomingVertex.Push)
	this.processVertexDependency = storage.NewDependencyNotifier(this.dagStorage.chanSettledVertex.Push)


	// Run every processing threads
	go this.dagStorage.chanIncomingVertex.Listen(this.OnIncomingVertex)
	go this.dagStorage.chanSettledVertex.Listen(this.OnSettledVertex)
	go this.dagStorage.chanSettledQueen.Listen((this.OnSettleQueen))

}

func (this *DagEngine) Start() {

	log.I("[dag] Begin DagEngine.Start()")

	// Load the previous cache data from Database: PendingPayloadData

	this.netProcessor.Start()
	log.I("[dag] End DagEngine.Start()")
}

func (this *DagEngine) Stop() {

}

//
func (this *DagEngine) SubmitPayload(data PayloadData) {

	log.I("[dag] Begin SubmitPayload.")

	this.dagStorage.queuePendingData.Push(data)

	if this.dagStorage.queuePendingData.IsFull() {

		// ComposePayloadVertex(nil)
		this.ComposeVertexEvent()
	}

	log.I("[dag] End SubmitPayload.")
}

func (this *DagEngine) ComposeVertexEvent() {

	log.I("[dag] Begin ComposeVertexEvent.")

	// Compose the Vertex Data
	vertex, _ := CreateVertex(this.dagStorage, this.dagNodes.GetSelf(), nil)

	if (vertex == nil) {
		log.W("Something wrong while generating new vertex, stopping composing.")
		return
	}

	// Handle the event
	if this.payloadHandler != nil {
		for _, payloadData := range vertex.GetContent().Data {
			this.payloadHandler.OnPayloadSubmitted(payloadData)
		}
	}

	// Send the Vertex to 0-2 nodes
	peerNodes := SelectPeerNodeToSendVertex(this.dagStorage, this.dagNodes)
	if peerNodes != nil && len(peerNodes) != 0 {

		for _, peerNode := range peerNodes {

			// Compose Vertex event and send
			vertexList, _ := FindPossibleUnknownVertexesForNode(this.dagStorage, peerNode)
			event, _ := ComposeVertexEvent(101, vertexList)
			data, _ := proto.Marshal(event)

			resultChan := this.netProcessor.SendEventToDeviceAsync(peerNode.Device, data)
			result := <-resultChan

			if (result.Err != nil) {
				log.I2("SendEvent finished. Result: eventId=[%d], err=[%s]", result.EventId, result.Err.Error())
			} else {
				log.I2("SendEvent succeeeded. Result: eventId=[%d]", result.EventId)
			}
		}
	}

	log.I("[dag] End ComposeVertexEvent.")
}

// Threads from Workers: push the incoming vertexes into channel
func (this *DagEngine) PushIncomingVertex(vertex *DagVertex) {

	log.I("PushIncomingVertex: push vertex: ", vertex.Hash)

	this.dagStorage.chanIncomingVertex.PushProto(vertex)
}

// Thread 1: Validate the incoming vertexes and build the dag
func (this *DagEngine) OnIncomingVertex(data []byte) {

	log.I("Begin OnIncomingVertex")

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

	log.I("End OnIncomingVertex")
}

// Thread 2: Mark vertex levels, and decide the candidates
func (this *DagEngine) OnSettledVertex(hash []byte) {

	log.I("Begin OnSettledVertex")

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
	log.I("End OnSettledVertex")
}

// Thread 3: Candidates vote for Queen and Decide Queen
func (this *DagEngine) OnSettledCandidate(candidateHash []byte) {

	log.I("Begin OnSettledCandidate")

	// The Collect vote method will decide new queens, and write into the chanSettledQueen
	ProcessCandidateVote(this.dagStorage, this.dagNodes, candidateHash, func(queenHash []byte) {
		this.dagStorage.chanSettledQueen.Push(queenHash)
	})

	log.I("End OnSettledCandidate")
}

// Thread 4: Queen to decide accept/reject vertex
func (this *DagEngine) OnSettleQueen(queenHash []byte) {

	log.I("Begin OnSettleQueen")

	completed := ProcessQueenDecision(this.dagStorage, queenHash)

	if completed == ProcessResult_Yes {

		} else {

			// If something wrong happened in processing, put the queen back to the queue and try again
			this.dagStorage.chanSettledQueen.Push(queenHash)
		}

	log.I("End OnSettleQueen")
}

