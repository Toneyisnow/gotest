package dag

import (
	"github.com/gogo/protobuf/proto"
	"github.com/smartswarm/go/log"
	"golang.org/x/crypto/openpgp/errors"
)
import "../network"

const (
	DAG_EVENT_QUEUE_MAX_LENGTH = 100
	DAG_EVENT_HANDLER_WORKER_COUNT = 10
)

type DagEventHandler struct {

	dagEngine    *DagEngine
	eventQueue   chan *DagEvent
	eventWorkers []*DagEventWorker
}

func NewDagEventHandler(engine *DagEngine) DagEventHandler {

	handler := DagEventHandler{}
	handler.eventQueue = make(chan *DagEvent, DAG_EVENT_QUEUE_MAX_LENGTH)
	handler.dagEngine = engine
	handler.eventWorkers = make([]*DagEventWorker, 0)

	return handler
}

func (this DagEventHandler) HandleEventData(context *network.NetContext, rawData []byte) (err error) {

	log.I("[dag][event handler] handling event data.")

	if this.dagEngine.EngineStatus != DagEngineStatus_Connected {
		log.W("[dag][event handler] dag engine is not in connected status, ignore this incoming event.")
	}

	dagEvent := new(DagEvent)
	ierr := proto.Unmarshal(rawData, dagEvent)
	if ierr != nil || dagEvent == nil {
		err = ierr
		return
	}

	// Put the data into queue, if the queue is reached max length, do something error
	isAllWorkersBusy := true
	for _, worker := range this.eventWorkers {
		if !worker.IsBusy() {
			isAllWorkersBusy = false
			break
		}
	}

	// Create new worker if needed
	if isAllWorkersBusy && len(this.eventWorkers) < DAG_EVENT_HANDLER_WORKER_COUNT {

		worker := NewDagEventWorker(this.eventQueue, this.dagEngine)
		this.eventWorkers = append(this.eventWorkers, worker)
		go worker.Start()
	}

	if len(this.eventQueue) < DAG_EVENT_QUEUE_MAX_LENGTH {
		this.eventQueue <- dagEvent
	} else {

		// Report error here
		err = errors.UnsupportedError("Queue is full and cannot process the event.")
	}

	return
}

func (this *DagEventHandler) StopAllWorkers() {

	for _, worker := range this.eventWorkers {

		worker.Stop()
	}
}