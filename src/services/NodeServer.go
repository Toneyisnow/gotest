package services

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"github.com/gorilla/websocket"
)
var upgrader = websocket.Upgrader{} // use default options

type NodeServer struct {

	ServerPort int
	AvailableNodeList []NodeServer

	httpServer *http.Server
	config *NodeConfig
}

func (this *NodeServer) Initialize() {

	this.config = LoadConfigFromFile()
}

func (this *NodeServer) Start() {

	var serveraddress = "localhost:" + strconv.Itoa(this.config.ServerPort)
	var serveraddr = flag.String("serveraddr", serveraddress, "http service serveraddress")

	// Start the server
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/events", handleEventMessage)

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

func handleEventMessage(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("recv: %s %s", message, mt)
	}
}

