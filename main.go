package main

import (
	"../gotest/swarm/dag"
	"../gotest/swarm/network"
	"../gotest/swarm/storage"
	"github.com/smartswarm/go/log"
	"os"
	"strconv"
	"time"
)

func main() {

	db_test()

	dag_test()

	time.Sleep(5 * time.Second)
}

func network_test() {

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

func dag_test() {

	handler := SamplePayloadHandler{}
	engine := dag.ComposeDagEngine(&handler)

	engine.Start()

	time.Sleep(5 * time.Second)
	for i :=  0; i < 15; i++ {

		data := "" + strconv.Itoa(i)
		engine.SubmitPayload([]byte(data))

		time.Sleep(1 * time.Second)
	}
}

func db_test() {

	storage := storage.ComposeRocksDBInstance("dag_test")
	storage.Put([]byte("key1"), []byte("111"))
	storage.Put([]byte("key2"), []byte("222"))
	storage.PutSeek([]byte("key1"), []byte("111"))
	storage.PutSeek([]byte("key2"), []byte("222"))

	data, _ := storage.Get([]byte("key1"))
	log.I("data: ", data)

	storage.Seek([]byte("key"), func(v []byte) {
		log.I("value", string(v))
	})
}

type SampleEventHandler struct {

}

func (this SampleEventHandler) HandleEventData(context *network.NetContext, rawData []byte) (err error) {

	log.I("HandleEventData Start.")

	return nil
}

type SamplePayloadHandler struct {

}

func (this *SamplePayloadHandler) OnPayloadSubmitted(data dag.PayloadData) {

	log.I("OnPayloadSubmitted: ", data)
}

func (this *SamplePayloadHandler) OnPayloadAccepted(data dag.PayloadData) {

}

func (this *SamplePayloadHandler) OnPayloadRejected(data dag.PayloadData) {

}
