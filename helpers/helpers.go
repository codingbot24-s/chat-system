package helpers

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)


type Clients []map[string]*websocket.Conn 

var (
	Connection Clients
	mu sync.Mutex
)


func addConnection(username string, conn *websocket.Conn) {
	mu.Lock()
	Connection = append(Connection, map[string]*websocket.Conn{username: conn})
	defer mu.Unlock()
}

func writeAll(ty int,msg []byte) {
	for _, m  := range Connection {
		for _, conn := range m {
			if err := conn.WriteMessage(ty,msg); err != nil {
				log.Printf("error writing to the client %w",err);
				return
			}
		}
	} 
}


func HandleConnection(conn *websocket.Conn) {

	// NOTE: (saad) we need to get the username for connection from the routes
	addConnection("test",conn);
	for {
		mty, b, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error reading from connection %w", err)
			break
		}
		fmt.Printf("recv  message %s\n", string(b))
		writeAll(mty,b)		
		// if err := conn.WriteMessage(msgt,b); err != nil {
		// 	log.Printf("error reading from connection %w", err)
		// 	break
		// }
	}

}
