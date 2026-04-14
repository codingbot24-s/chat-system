package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/codingbot24.s/chat-system/db"
	"github.com/codingbot24.s/chat-system/handlers"
	"github.com/gorilla/mux"
)


func main() {
	db := db.Init();
	h := handlers.NewHandler(db)
	r := mux.NewRouter();
	
	r.HandleFunc("/health",h.Health).Methods("GET");
	r.HandleFunc("/ws",h.HandleWebSocket).Methods("Get")
	r.HandleFunc("/signup",h.SignUp).Methods("POST")





	fmt.Println("starting the http server on 8000");
	if err := http.ListenAndServe(":8000",r); err != nil {
		log.Fatalf("cant start the server %w",err);
	}
}
