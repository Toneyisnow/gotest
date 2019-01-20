package network

import (
	"errors"
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

	connection  *websocket.Conn
	_serverPort int32

	clientServerHostUrl string
}

type ConnectionCallback func(socket *NetWebSocket)


func (wc *NetWebSocket) ReadMessage() (message *NetMessage, err error) {

	mt, rawMessageBytes, re := wc.connection.ReadMessage()
	log.I("[network] websocket got read message.")
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

	log.I("[network] web socket write message: type=", message.MessageType, ";id=", message.MessageId)

	bytes, err := proto.Marshal(message)

	n = len(bytes)
	err = wc.connection.WriteMessage(websocket.TextMessage, bytes)
	return
}

func (wc *NetWebSocket) Close() error {
	return wc.connection.Close()
}

func (wc *NetWebSocket) LocalHostAddress() net.Addr {
	return wc.connection.LocalAddr()
}

func (wc *NetWebSocket) RemoteHostAddress() net.Addr {
	return wc.connection.RemoteAddr()
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

		socket := &NetWebSocket{connection: conn, _serverPort:serverPort}

		// Get the client ServerHostUrl
		_, rawMessageBytes, _ := conn.ReadMessage()
		log.I("[network] raw message bytes: ", string(rawMessageBytes))
		socket.clientServerHostUrl = string(rawMessageBytes)

		handler(socket)
	})

	server := &http.Server{Addr: addr, Handler: serveMutex}

	if err := server.ListenAndServe(); err == http.ErrServerClosed {
		log.I("[network] serve closed")
	} else {
		log.W("[network] ======== serve failed! ========", err)
	}
}

func StartWebSocketDial(device *NetDevice, clientServerHostUrl string) (socket *NetWebSocket, err error) {

	retryCount := 3
	retryInterval := time.Second

	var peeraddress = device.GetHostUrl()
	// var peeraddr = flag.String("peeraddr", peeraddress, "http service peeraddress")
	peeraddr := string(peeraddress)
	u := url.URL{Scheme: "ws", Host: peeraddr, Path: "/"}

	for i := 0; i < retryCount; i++ {

		_conn, _, de := websocket.DefaultDialer.Dial(u.String(), nil);
		if (de == nil) {

			log.W("[network] connect succeeded.")

			// Send the clientServerHostUrl
			de = _conn.WriteMessage(websocket.TextMessage, []byte(clientServerHostUrl))
			socket = &NetWebSocket{connection: _conn}

			err = nil
			return
		}

		err = de
		time.Sleep(retryInterval)
	}

	// log.W("[network] connect failed!", err)

	return
}
