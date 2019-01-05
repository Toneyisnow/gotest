package main

import (
	"../gotest/swarm/dag"
	"../gotest/swarm/network"
	"../gotest/swarm/storage"
	"encoding/binary"
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

	time.Sleep(3 * time.Second)
	for i :=  0; i < 15; i++ {

		data := "" + strconv.Itoa(i)
		engine.SubmitPayload([]byte(data))

		time.Sleep(1 * time.Second)
	}
}

func db_test() {

	rstorage := storage.ComposeRocksDBInstance("dag_test")

	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, 111)

	rstorage.Put(append([]byte("key1"), bs...), []byte("111"))
	rstorage.Put([]byte("key2"), []byte("222"))
	rstorage.PutSeek([]byte("key"), []byte("111"))
	rstorage.PutSeek([]byte("key2"), []byte("222"))
	rstorage.PutSeek([]byte("key3"), []byte("333"))
	rstorage.PutSeek([]byte("kwy4"), []byte("444"))
	rstorage.PutSeek([]byte("key5"), []byte("555"))
	rstorage.PutSeek([]byte("key11"), []byte("11555"))

	data, _ := rstorage.Get([]byte("key1"))
	log.I("data: ", data)

	rstorage.SeekAll([]byte("key2"), func(v []byte) {
		log.I("value", string(v))
	})

	levelQueue := storage.ComposeRocksLevelQueue(rstorage, "incomingVertexQueue")
	levelQueue.Push(1, []byte("83291"))

	result := levelQueue.Pop()
	log.I("LevelQueue: Pop result: ", result)

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
