package network

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"github.com/smartswarm/go/log"
	"networking/pb"
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
	if mt, rawMessageBytes, re := wc._connection.ReadMessage(); re != nil {
		err = re
	} else if mt != websocket.TextMessage {
		err = e_unsupported_message_type
	} else {

		message := new(pb.NetMessage)
		proto.Unmarshal(rawMessageBytes, message)
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
