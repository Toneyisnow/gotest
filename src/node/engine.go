package node

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
	"net/http"
	"strconv"
	"github.com/gorilla/websocket"
)

type Engine struct {
	ServerPort int
	PeerPort int
}

var addrOld = flag.String("addrOld", "localhost:8080", "http service address")
var upgrader = websocket.Upgrader{} // use default options

func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	fmt.Println(r)
	return string(r)
}

func echo(w http.ResponseWriter, r *http.Request) {
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
		
		/*
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
		*/
	}
}

func (engine Engine) Start(s string) {

	var serveraddress = "localhost:" + strconv.Itoa(engine.ServerPort)
	var serveraddr = flag.String("serveraddr", serveraddress, "http service serveraddress")
	var peeraddress = "localhost:" + strconv.Itoa(engine.PeerPort)
	var peeraddr = flag.String("peeraddr", peeraddress, "http service peeraddress")
	
	// Start the server
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	go http.ListenAndServe(*serveraddr, nil)
	
	// Start the client
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *peeraddr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	for {
		if err != nil {
			log.Printf("dial:", err)
			log.Printf("wait for 3 seconds...")
			
			time.Sleep(3 * time.Second)
			c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
			log.Printf("Dialed again...")
			
		} else {
			break
		}
	}
	
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	
	
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	
	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("Test message..."))
			if err != nil {
				log.Println("write:", err, t)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "This is test."))
			// err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "This is test."))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}