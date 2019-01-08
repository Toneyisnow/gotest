package dag

import (
	"../network"
	"github.com/gogo/protobuf/proto"
	"github.com/smartswarm/go/log"
	"os"
	"strconv"
)

type PayloadData []byte

const (
	DAG_PAYLOAD_BUFFER_SIZE = 10
)

type DagEngine struct {

	payloadHandler PayloadHandler

	// payloadDataQueueMutex    sync.Mutex
	// pendingPayloadDataQueues []PayloadData

	dagNodes    *DagNodes
	netTopology *network.NetTopology

	dagStorage *DagStorage

	// Stage 1: IncomingVertexQueue: collect and save the vertexes that received from other nodes
	incomingVertexQueue chan *DagVertex

	// Stage 2: FreshVertexQueue: the vertexes that saved in topology order waiting to be processed
	freshVertexQueue chan []byte

	// Stage 3: FreshCandidateQueue: the candidates that waiting to be processed
	freshCandidateQueue chan []byte

	// Stage 4: QueenQueue: the queue of the queens that could confirm the rest of the vertexes Accepted/Rejected
	queenQueue chan []byte

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
	this.incomingVertexQueue = make(chan *DagVertex)
	this.freshVertexQueue = make(chan []byte)
	this.freshCandidateQueue = make(chan []byte)
	this.queenQueue = make(chan []byte)


	this.dagStorage = DagStorageGetInstance()

	eventHandler := NewDagEventHandler(this)
	this.netProcessor = network.CreateProcessor(this.netTopology, eventHandler)

	// Run every processing threads
	go this.ProcessIncomingVertexThread()
	go this.ProcessVertexInDagThread()
	go this.ProcessCandidateVoteThread()
	go this.ProcessQueenDecisionThread()

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

	if this.dagStorage.queuePendingData.DataSize() >= DAG_PAYLOAD_BUFFER_SIZE {

		// ComposePayloadVertex(nil)
		this.ComposeVertexEvent()
	}

	log.I("[dag] End SubmitPayload.")
}

func (this *DagEngine) ComposeVertexEvent() {

	log.I("[dag] Begin ComposeVertexEvent.")

	// Compose the Vertex Data
	vertex, _ := CreateVertex(this.dagNodes.GetSelf(), nil)

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

	// Send the Vertex to 0-1 nodes
	peerNode := SelectPeerNodeToSendVertex(this.dagStorage, this.dagNodes)
	if peerNode != nil {

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

	log.I("[dag] End ComposeVertexEvent.")
}

// Threads from Workers: push the incoming vertexes into queue
func (this *DagEngine) PushIncomingVertex(vertex *DagVertex) {

	log.I("PushIncomingVertex: push vertex: ", vertex.Hash)
	// Save to storage
	this.dagStorage.queueIncomingVertex.PushProto(vertex)

	// Push to channel to notify processor
	this.incomingVertexQueue <- vertex
}

// Thread 1: Validate the incoming vertexes and build the dag
func (this *DagEngine) ProcessIncomingVertexThread() {

	log.I("Begin ProcessIncomingVertexThread")
	for {
		incomingVertex := <-this.incomingVertexQueue

		log.I("ProcessIncomingVertexThread: got incomingVertex from channel. Hash=", incomingVertex.Hash)
		this.dagStorage.queueIncomingVertex.Pop()

		decision := ProcessIncomingVertex(this.dagStorage, incomingVertex)

		switch decision {
		case ProcessResult_Yes:
			this.freshVertexQueue <- incomingVertex.Hash
			break;
		case ProcessResult_No:
			// Discard this vertex if it's invalid
			break;
		case ProcessResult_Undecided:
			// If it's undecided, put it back into the incoming queue
			this.PushIncomingVertex(incomingVertex)
			break;
		}
	}
}

// Thread 2: Mark vertex levels, and decide the candidates
func (this *DagEngine) ProcessVertexInDagThread() {

	for {
		vertexHash := <-this.freshVertexQueue

		log.I("ProcessVertexInDagThread: got vertex from channel. Hash=", vertexHash)
		this.dagStorage.queueIncomingVertex.Pop()

		result := ProcessVertexAndDecideCandidate(this.dagStorage, vertexHash)
		log.I("ProcessVertexInDagThread: result=", result)

		switch result {
		case ProcessResult_Yes:
			this.dagStorage.queueCandidate.Push(vertexHash)
			this.freshCandidateQueue <- vertexHash
			break;
		case ProcessResult_No:
		case ProcessResult_Undecided:
			break;
		}
	}
}

// Thread 3: Vote for Queen
func (this *DagEngine) ProcessCandidateVoteThread() {

	for {
		candidateHash := <-this.freshCandidateQueue

		log.I("ProcessCandidateVoteThread: got candidate from channel. Hash=", candidateHash)
		this.dagStorage.queueCandidate.Pop()

		ProcessCandidateVote(this.dagStorage, candidateHash)
		result := ProcessCandidateCollectVote(this.dagStorage, candidateHash)
		log.I("ProcessVertexInDagThread: result=", result)

		if result == ProcessResult_Yes {
			// Found new queen
			for {
				queenHash := this.dagStorage.queueQueen.Pop()
				if queenHash == nil {
					break
				}

				this.queenQueue <- queenHash
			}
		}
	}
}

// Thread 4: Queen to decide accept/reject vertex
func (this *DagEngine) ProcessQueenDecisionThread() {

	for {
		queenHash := <-this.queenQueue
		this.dagStorage.queueQueen.Pop()

		result := ProcessQueenDecision(this.dagStorage, queenHash)

		if result == ProcessResult_Yes {


		} else {

			// If something wrong happened in processing, put the queen back to the queue and try again
			this.queenQueue <- queenHash
		}
	}
}

