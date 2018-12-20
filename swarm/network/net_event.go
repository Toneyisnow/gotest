package network

import "math/rand"

type NetEvent struct {

	EventId uint32
	Data []byte

}

type NetEventResult struct {

	EventId uint32
	Err error
}


func ComposeEvent(data []byte) *NetEvent {

	event := new(NetEvent)

	event.EventId = rand.Uint32()
	event.Data = data

	return event
}

func ComposeEventResult(event *NetEvent, err error) *NetEventResult {

	result := new(NetEventResult)

	result.EventId = event.EventId
	result.Err = err

	return result
}
