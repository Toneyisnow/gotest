package main

import (
	"../gotest/swarm/dag"
	"../gotest/swarm/network"
	"../gotest/swarm/storage"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/smartswarm/core/crypto/secp256k1"
	"io/ioutil"

	//"../gotest/swarm/storagetest"

	"encoding/binary"
	"github.com/smartswarm/go/log"
	"os"
	"strconv"
	"time"
)

func main() {

	// test_args()
	// generate_signatures()

	// db_test()

	dag_test()

	time.Sleep(5 * time.Second)
}

func crpto_test() {

	// sha := sha256.New()
	hash := sha256.Sum256([]byte("Hello world"))
	// hashString := base64.URLEncoding.EncodeToString([]byte(hash))


	pubKey, privKey := secp256k1.GenerateKeyPair()


	str := hex.EncodeToString(hash[:])
	log.I(str)

	fmt.Println("Hash: ", )
	fmt.Println("Public Key: ", hex.EncodeToString(pubKey[:]))
	log.I("Public Key:", )
	log.I("Private Key:", string(privKey))

	// secp256k1.Sign()


}

func test_args() {


	fmt.Println("os.Args len:", len(os.Args))

	fileName := string(os.Args[1])
	fmt.Println(fileName)
}

func generate_signatures() {

	jsonFile, err := os.Open("node-topology.json")

	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)
	dagNodes := new(dag.DagNodes)
	json.Unmarshal(byteValue, dagNodes)

	jsonFile.Close()


	pubKey, privKey := secp256k1.GenerateKeyPair()

	dagNodes.Self.Device.PublicKey = pubKey
	dagNodes.Self.Device.PrivateKey = privKey

	for _, n := range dagNodes.Peers {

		pubKey, privKey = secp256k1.GenerateKeyPair()
		n.Device.PublicKey = pubKey
		n.Device.PrivateKey = privKey
	}

	dagNodes.SaveToFile("node-topology-2.json")
}

func network_test() {

	topology := network.LoadTopology()

	if (len(os.Args) > 1) {
		serverPort, _ := strconv.Atoi(os.Args[1])
		topology.Self().Port = int32(serverPort)
	}

	eventHandler := SampleEventHandler{}

	netProcessor := network.CreateProcessor(topology, eventHandler)

	netProcessor.StartServer()

	time.Sleep(2000)

	// Send some event data
	for _, device := range topology.GetAllRemoteDevices() {

		data := []byte("A Sample Event")

		log.I2("SendEvent started: eventData:[%s]", data)

		result := <- netProcessor.SendEventToDeviceAsync(device, data)

		if (result.Err != nil) {
			log.I2("SendEvent finished. Result: eventId=[%d], err=[%s]", result.EventId, result.Err.Error())
		} else {
			log.I2("SendEvent succeeeded. Result: eventId=[%d]", result.EventId)
		}
	}
}

func dag_test() {

	if len(os.Args) < 2 {
		fmt.Println("Missing config file in command. Usage: [_exe_] <ConfigFile.json>")
		return
	}

	configFileName := string(os.Args[1])
	config := dag.LoadConfigFromJsonFile(configFileName)

	handler := SamplePayloadHandler{}
	engine := dag.NewDagEngine(config, &handler)

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

	sQueue := storage.NewRocksSequenceQueue(rstorage, "iiiQueue", 100)
	sQueue.Push([]byte("111"))
	sQueue.Push([]byte("222"))
	sQueue.Push([]byte("333"))

	result := sQueue.Pop()
	log.I("Pop result: ", result)
	result = sQueue.Pop()
	log.I("Pop result: ", result)
	result = sQueue.Pop()
	log.I("Pop result: ", result)
	result = sQueue.Pop()
	log.I("Pop result: ", result)


	table := storage.NewRocksTable(rstorage, "sampleTable")
	table.InsertOrUpdate([]byte("key1"), []byte("111"))

	val := table.Get([]byte("key1"))

	log.I("Table got value: ", val)
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

	log.I("OnPayloadAccepted: ", data)
}

func (this *SamplePayloadHandler) OnPayloadRejected(data dag.PayloadData) {

	log.I("OnPayloadRejected: ", data)
}
