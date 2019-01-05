package dag

import (
	"../network"
	"github.com/gogo/protobuf/proto"
	"github.com/smartswarm/go/log"
	"os"
	"strconv"
	"sync"
)

type PayloadData []byte

const (
	DAG_PAYLOAD_BUFFER_SIZE = 10
)

type DagEngine struct {

	_payloadHandler PayloadHandler

	_payloadDataQueueMutex sync.Mutex
	_pendingPayloadDataQueue []PayloadData

	_dagNodes *DagNodes
	_topology *network.NetTopology

	_dagStorage *DagStorage

	_netProcessor *network.NetProcessor
}

func ComposeDagEngine(handler PayloadHandler) *DagEngine {

	engine := new(DagEngine)
	engine.BindHandler(handler)
	engine.Initialize()

	return engine
}


func (this *DagEngine) BindHandler(handler PayloadHandler) {

	this._payloadHandler = handler
}

func (this *DagEngine) Initialize() {

	this._pendingPayloadDataQueue = make([]PayloadData, 0)

	this._dagNodes = LoadDagNodesFromFile("node-topology.json")
	this._topology = this._dagNodes.GetNetTopology()

	if (len(os.Args) > 1) {
		serverPort, _ := strconv.Atoi(os.Args[1])
		this._topology.Self().Port = int32(serverPort)
	}

	this._dagStorage = DagStorageGetInstance()

	eventHandler := ComposeDagEventHandler(this)
	this._netProcessor = network.CreateProcessor(this._topology, eventHandler)

}

func (this *DagEngine) Start() {

	log.I("[dag] Begin DagEngine.Start()")

	// Load the previous cache data from Database: PendingPayloadData


	this._netProcessor.Start()

	log.I("[dag] End DagEngine.Start()")

}

func (this *DagEngine) Stop() {

}

//
func (this *DagEngine) SubmitPayload(data PayloadData) {

	log.I("[dag] Begin SubmitPayload.")

	this._payloadDataQueueMutex.Lock()

	this._pendingPayloadDataQueue = append(this._pendingPayloadDataQueue, data)
	this._dagStorage.PutPendingPayloadData(data)

	this._payloadDataQueueMutex.Unlock()

	if len(this._pendingPayloadDataQueue) >= DAG_PAYLOAD_BUFFER_SIZE {

		// ComposePayloadVertex(nil)
		this.ComposeVertexEvent()
	}

	log.I("[dag] End SubmitPayload.")
}

func (this *DagEngine) ComposeVertexEvent() {

	log.I("[dag] Begin ComposeNewVertex.")

	// Compose the Vertex Data
	vertex, _ := GenerateNewVertex(this._dagNodes.GetSelf(), nil)

	if (vertex == nil) {
		log.W("Something wrong while generating new vertex, stopping composing.")
		return
	}

	// Send the Vertex to some nodes
	for _, node := range this._dagNodes.Peers {

		vertexList, _ := FindPossibleUnknownVertexesForNode(node)

		event, _ := ComposeVertexEvent(101, vertexList)
		data, _ := proto.Marshal(event)

		resultChan := this._netProcessor.SendEventToDeviceAsync(node.Device, data)
		result := <-resultChan

		if (result.Err != nil) {
			log.I2("SendEvent finished. Result: eventId=[%d], err=[%s]", result.EventId, result.Err.Error())
		} else {
			log.I2("SendEvent succeeeded. Result: eventId=[%d]", result.EventId)
		}
	}

	// Handle the event
	if this._payloadHandler != nil {

		for _, payloadData := range this._pendingPayloadDataQueue {
			this._payloadHandler.OnPayloadSubmitted(payloadData)
		}
	}

	// Clear the queue
	this._pendingPayloadDataQueue = make([]PayloadData, 0)

	log.I("[dag] End ComposeNewVertex.")
}