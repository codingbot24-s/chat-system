package routes

import (
	"github.com/codingbot24.s/chat-system/handlers"
	"github.com/gorilla/mux"
)


func StartRouter (r* mux.Router) {
	r.HandleFunc("/",handlers.Health).Methods("GET");
	r.HandleFunc("/ws",handlers.HandleWebSocket)
}