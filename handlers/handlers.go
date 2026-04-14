package handlers

import (
	"log"
	"net/http"

	"github.com/codingbot24.s/chat-system/helpers"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgarding con %w", err)
		return
	}
	go helpers.HandleConnection(conn)
}
