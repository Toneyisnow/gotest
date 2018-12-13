package services

import (
	"flag"
	"github.com/czsilence/gorocksdb"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	// "objectmodels/network"

	// "io/ioutil"
	"log"
	"net/http"
	"networking/pb"
	"strconv"
	"time"
)
var upgrader = websocket.Upgrader{} // use default options

type NodeServer struct {

	ServerPort int

	httpServer *http.Server
	config *NodeConfig

	messageQueue chan *pb.NetMessage
}

func (this *NodeServer) Initialize(msgQueue chan *pb.NetMessage) {

	this.config = LoadConfigFromFile()
	this.messageQueue = msgQueue
}

func (this *NodeServer) Start() {

	var serveraddress = "localhost:" + strconv.Itoa(this.config.Self.ServerPort)
	var serveraddr = flag.String("serveraddr", serveraddress, "http service serveraddress")

	// Start the server
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/events", this.handleEventMessage)

	this.httpServer = &http.Server{Addr: *serveraddr}

	go func() {
		if err := this.httpServer.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

}

func (server *NodeServer) Stop() {

	if err := server.httpServer.Shutdown(nil); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
}

func (this *NodeServer) handleEventMessage(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		mess := new(pb.NetMessage)
		proto.Unmarshal(message, mess)

		log.Printf("recv: OwnerId=%s Hash=%s %s", mess.OwnerId, mess.Hash, mt)

		if (mess.Type == pb.NetMessageType_SendEvents) {
			log.Printf("recv: EventId=%s %s", mess.SendEventMessage.Events[0].EventId, mt)

		}

		this.messageQueue <- mess
	}
}

func (this *NodeServer) handleEventMessage2(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	db := newTestDB("TestDBGet", nil)
	defer db.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		mess := new(pb.NetMessage)
		proto.Unmarshal(message, mess)

		log.Printf("recv: OwnerId=%s Hash=%s %s", mess.OwnerId, mess.Hash, mt)

		if (mess.Type == pb.NetMessageType_SendEvents) {
			log.Printf("recv: EventId=%s %s", mess.SendEventMessage.Events[0].EventId, mt)

		}

		this.messageQueue <- mess

		var (
			givenKey  = []byte(mess.OwnerId)
			givenVal1 = []byte(mess.Hash)
			wo        = gorocksdb.NewDefaultWriteOptions()
			ro        = gorocksdb.NewDefaultReadOptions()
		)

		// create
		db.Put(wo, givenKey, givenVal1)

		time.Sleep(2 * time.Second)

		// retrieve
		v1, _ := db.Get(ro, givenKey)
		log.Printf("db Get: %s %s", givenKey, v1.Data())


		defer v1.Free()
	}
}


func newTestDB(name string, applyOpts func(opts *gorocksdb.Options)) *gorocksdb.DB {
	//dir, _ := ioutil.TempDir("", "gorocksdb-"+name)


	dir := "/var/folders/gorocsdb-Test"
	log.Printf("tempDir: %s", dir)
	// ensure.Nil(t, err)

	opts := gorocksdb.NewDefaultOptions()
	// test the ratelimiter
	rateLimiter := gorocksdb.NewRateLimiter(1024, 100*1000, 10)
	opts.SetRateLimiter(rateLimiter)
	opts.SetCreateIfMissing(true)
	if applyOpts != nil {
		applyOpts(opts)
	}
	db, _ := gorocksdb.OpenDb(opts, dir)
	// ensure.Nil(t, err)

	return db
}


