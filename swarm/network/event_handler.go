package network

type EventHandler interface {

	// This is the main function for the upper level to handle the received message data
	HandleEventData(context *NetContext, rawData []byte) (err error)
}
