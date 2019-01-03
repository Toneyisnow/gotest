package dag

import "strconv"

type DagEventExtension struct {

}

func ComposeVertexEvent(eventId int, vertexList []*DagVertex) (event *DagEvent, err error) {

	event = new(DagEvent)
	event.EventId = strconv.Itoa(eventId)
	event.EventType = DagEventType_VertexesData

	vertexesEvent := new(VertexesDataEvent)
	vertexesEvent.Vertexes = make([]*DagVertex, 0)

	for _, vertex := range vertexList {
		vertexesEvent.Vertexes = append(vertexesEvent.Vertexes, vertex)
	}

	event.Data = &DagEvent_VertexesDataEvent{VertexesDataEvent: vertexesEvent}

	err = nil
	return;
}

func ComposeInfoEvent() (event *DagEvent, err error) {

	event = new(DagEvent)

	return;
}
