package main

import (
	"../gotest/swarm/network"
	"fmt"
	"time"
)

func main() {

	sample_test()

	time.Sleep(10 * time.Second)
}

func sample_test() {

	topology := network.LoadTopologyFromDB("")

	eventHandler := SampleEventHandler{}

	netProcessor := network.CreateProcessor(topology, eventHandler)

	netProcessor.Start()
}

type SampleEventHandler struct {

}

func (this SampleEventHandler) HandleEventData(context *network.NetContext, rawData []byte) {

	fmt.Printf("HandleEventData")
}