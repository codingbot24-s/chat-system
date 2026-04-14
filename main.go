package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/codingbot24.s/chat-system/routes"
	"github.com/gorilla/mux"
)


func main() {
	r := mux.NewRouter();
	routes.StartRouter(r);

	fmt.Println("starting the http server on 8000");
	if err := http.ListenAndServe(":8000",r); err != nil {
		log.Fatalf("cant start the server %w",err);
	}
	
}
