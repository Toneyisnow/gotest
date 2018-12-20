package main

import (
	"../gotest/swarm/network"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {

	sample_test()

	time.Sleep(10 * time.Second)
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
}

type SampleEventHandler struct {

}

func (this SampleEventHandler) HandleEventData(context *network.NetContext, rawData []byte) {

	fmt.Printf("HandleEventData")
}