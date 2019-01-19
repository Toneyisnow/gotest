package dag

import (
	"math/rand"
	"strconv"
)
/*
// Deprecated?
func NewVertexEvent(eventId int, vertexList []*DagVertex) (event *DagEvent, err error) {

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
*/

func NewVertexEvent(mainVertex *DagVertex, relatedVertexes []*DagVertex) (event *DagEvent, err error) {

	event = new(DagEvent)
	event.EventId = strconv.Itoa(rand.Intn(100000) + 100000)
	event.EventType = DagEventType_VertexesData

	vertexesEvent := new(VertexesDataEvent)

	vertexesEvent.MainVertex = mainVertex
	vertexesEvent.Vertexes = make([]*DagVertex, 0)
	for _, vertex := range relatedVertexes {
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
