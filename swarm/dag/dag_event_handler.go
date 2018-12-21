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

	_engine *DagEngine
	_eventQueue chan *DagEvent
	_eventWorkers []*DagEventWorker
}

func ComposeDagEventHandler(engine *DagEngine) DagEventHandler {

	handler := DagEventHandler{}
	handler._eventQueue = make(chan *DagEvent, DAG_EVENT_QUEUE_MAX_LENGTH)
	handler._engine = engine
	handler._eventWorkers = make([]*DagEventWorker, 0)

	return handler
}

func (this DagEventHandler) HandleEventData(context *network.NetContext, rawData []byte) (err error) {

	log.I("HandleEventData Start.")

	dagEvent := new(DagEvent)
	ierr := proto.Unmarshal(rawData, dagEvent)
	if ierr != nil || dagEvent == nil {
		err = ierr
		return
	}

	// Put the data into queue, if the queue is reached max length, do something error
	isAllWorkersBusy := true
	for _, worker := range this._eventWorkers {
		if !worker.IsBusy() {
			isAllWorkersBusy = false
			break
		}
	}

	// Create new worker if needed
	if isAllWorkersBusy && len(this._eventWorkers) < DAG_EVENT_HANDLER_WORKER_COUNT {

		worker := ComposeDagEventWorker(this._eventQueue, this._engine)
		this._eventWorkers = append(this._eventWorkers, worker)
		go worker.Start()
	}

	if len(this._eventQueue) < DAG_EVENT_QUEUE_MAX_LENGTH {
		this._eventQueue <- dagEvent
	} else {

		// Report error here
		err = errors.UnsupportedError("Queue is full and cannot process the event.")
	}

	return
}

func (this *DagEventHandler) StopAllWorkers() {

	for _, worker := range this._eventWorkers {

		worker.Stop()
	}
}