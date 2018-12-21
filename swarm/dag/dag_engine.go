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

	_topology *network.NetTopology

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

	this._topology = network.LoadTopology()

	if (len(os.Args) > 1) {
		serverPort, _ := strconv.Atoi(os.Args[1])
		this._topology.Self().Port = int32(serverPort)
	}

	eventHandler := ComposeDagEventHandler(this)
	this._netProcessor = network.CreateProcessor(this._topology, eventHandler)

}

func (this *DagEngine) Start() {

	log.I("[dag] Begin DagEngine.Start()")

	this._netProcessor.Start()

	log.I("[dag] End DagEngine.Start()")

}

func (this *DagEngine) Stop() {

}

//
func (this *DagEngine) SubmitPayload(data PayloadData) {

	log.I("[dag] Begin SubmitPayload.")
	this._pendingPayloadDataQueue = append(this._pendingPayloadDataQueue, data)

	if len(this._pendingPayloadDataQueue) >= DAG_PAYLOAD_BUFFER_SIZE {
		this.ComposeNewVertex()
	}

	log.I("[dag] End SubmitPayload.")
}

func (this *DagEngine) ComposeNewVertex() {

	log.I("[dag] Begin ComposeNewVertex.")

	this._payloadDataQueueMutex.Lock()
	defer this._payloadDataQueueMutex.Unlock()

	// Compose the Vertex Data

	event := new(DagEvent)
	event.EventId = "11"
	event.EventType = DagEventType_Info
	data, _ := proto.Marshal(event)

	for _, device := range this._topology.GetAllRemoteDevices() {

		resultChan := this._netProcessor.SendEventToDeviceAsync(device, data)
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