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

	this.dagStorage = DagStorageGetInstance()

	eventHandler := ComposeDagEventHandler(this)
	this.netProcessor = network.CreateProcessor(this.netTopology, eventHandler)

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