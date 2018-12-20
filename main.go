package main

import (
	"../gotest/swarm/network"
	"github.com/smartswarm/go/log"
	"os"
	"strconv"
	"time"
)

func main() {

	sample_test()

	time.Sleep(3 * time.Second)
}

func sample_test() {

	topology := network.LoadTopology()

	if (len(os.Args) > 1) {
		serverPort, _ := strconv.Atoi(os.Args[1])
		topology.Self().Port = int32(serverPort)
	}



	eventHandler := SampleEventHandler{}

	netProcessor := network.CreateProcessor(topology, eventHandler)

	netProcessor.Start()

	time.Sleep(2000)

	// Send some event data
	for _, device := range topology.GetAllRemoteDevices() {

		data := []byte("A Sample Event")

		log.I2("SendEvent started: eventData:[%s]", data)

		resultChan := netProcessor.SendEventToDeviceAsync(device, data)
		result := <- resultChan

		if (result.Err != nil) {
			log.I2("SendEvent finished. Result: eventId=[%d], err=[%s]", result.EventId, result.Err.Error())
		} else {
			log.I2("SendEvent succeeeded. Result: eventId=[%d]", result.EventId)
		}

	}

}

type SampleEventHandler struct {

}

func (this SampleEventHandler) HandleEventData(context *network.NetContext, rawData []byte) {

	log.I("HandleEventData Start.")
}