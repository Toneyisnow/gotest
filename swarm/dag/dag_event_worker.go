package dag

type DagEventWorker struct {

	_engine *DagEngine

	_isBusy bool
	_eventQueue chan *DagEvent

	_stopSignal chan int
}

func ComposeDagEventWorker(queue chan *DagEvent, engine *DagEngine) *DagEventWorker {

	worker := new(DagEventWorker)
	worker._isBusy = false
	worker._stopSignal = make(chan int)
	worker._eventQueue = queue
	worker._engine = engine

	return worker
}

func (this *DagEventWorker) DoEvent(event *DagEvent) {

	// Do the actual work here

	if event.GetEventType() == DagEventType_VertexesData {

		vertexesDataEvent := event.GetVertexesDataEvent()
		for _, vertex := range vertexesDataEvent.Vertexes {
			this._engine.PushIncomingVertex(vertex)
		}
	}
}

func (this *DagEventWorker) IsBusy() bool {
	return this._isBusy
}

func (this *DagEventWorker) Stop() {

	this._stopSignal <- 1
}

func (this *DagEventWorker) Start() {

	for {
		select {
			case event := <- this._eventQueue:
				this._isBusy = true
				this.DoEvent(event)
				this._isBusy = false
			case stop := <-this._stopSignal:
				if stop == 1 {
						return
				}
		}
	}
}
