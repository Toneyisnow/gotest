package network

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/smartswarm/go/log"
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	e_unsupported_message_type = errors.New("unsupported message type")
)

var wsUpgrader = websocket.Upgrader{}

type NetWebSocket struct {

	_connection *websocket.Conn
	_serverPort int32
}

type ConnectionCallback func(socket *NetWebSocket)


func (wc *NetWebSocket) ReadMessage() (message *NetMessage, err error) {

	log.I("[network] websocket read message.")

	mt, rawMessageBytes, re := wc._connection.ReadMessage()
	log.I("[network] raw message bytes: ", string(rawMessageBytes))

	if  re != nil {
		err = re

	} else if mt != websocket.TextMessage {
		err = e_unsupported_message_type
	} else {

		log.I("[network] raw message bytes: ", string(rawMessageBytes))
		message = new(NetMessage)
		pe := proto.Unmarshal(rawMessageBytes, message)
		if pe != nil {
			err = pe
		} else {
			return message, nil
		}
	}

	return
}

func (wc *NetWebSocket) Write(message *NetMessage) (n int, err error) {

	bytes, err := proto.Marshal(message)

	n = len(bytes)
	err = wc._connection.WriteMessage(websocket.TextMessage, bytes)
	return
}

func (wc *NetWebSocket) Close() error {
	return wc._connection.Close()
}

func (wc *NetWebSocket) LocalHostAddress() net.Addr {
	return wc._connection.LocalAddr()
}

func (wc *NetWebSocket) RemoteHostAddress() net.Addr {
	return wc._connection.RemoteAddr()
}

// 启动websocket服务器
func StartWebSocketListen(serverPort int32, handler ConnectionCallback) {

	if serverPort == 0 {
		log.I("[network] ignore listen for incoming connection.")
		return
	}

	var addr string = fmt.Sprintf("0.0.0.0:%d", serverPort)
	log.I("[network] ======== serve start, at:", addr)
	var serveMutex = http.NewServeMux()

	serveMutex.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// 移除跨域的header
		r.Header.Del("Origin")

		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.E("[network]", err)
			return
		}
		handler(&NetWebSocket{_connection: conn, _serverPort:serverPort})
	})

	server := &http.Server{Addr: addr, Handler: serveMutex}


	if err := server.ListenAndServe(); err == http.ErrServerClosed {
		log.I("[network] serve closed")
	} else {
		log.W("[network] ======== serve failed! ========", err)
	}
}

func StartWebSocketDial(device *NetDevice) (socket *NetWebSocket, err error) {

	retryCount := 5
	retryInterval := 3 * time.Second

	var peeraddress = device.GetHostUrl()
	var peeraddr = flag.String("peeraddr", peeraddress, "http service peeraddress")
	u := url.URL{Scheme: "ws", Host: *peeraddr, Path: "/"}

	for i := 0; i < retryCount; i++ {

		_conn, _, de := websocket.DefaultDialer.Dial(u.String(), nil);
		if (de == nil) {

			log.W("[network] connect succeeded.")
			socket = &NetWebSocket{_connection: _conn}
			err = nil
			return
		}

		err = de
		log.W("[network] connect failed, retrying", err)
		time.Sleep(retryInterval)
	}

	log.W("[network] connect failed!", err)

	return
}
