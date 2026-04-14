package helpers

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

type Clients []map[string]*websocket.Conn

var (
	Connection Clients
	mu         sync.Mutex
)

func addConnection(username string, conn *websocket.Conn) {
	mu.Lock()
	Connection = append(Connection, map[string]*websocket.Conn{username: conn})
	defer mu.Unlock()
}

func writeAll(ty int, msg []byte) {
	for _, m := range Connection {
		for _, conn := range m {
			if err := conn.WriteMessage(ty, msg); err != nil {
				log.Printf("error writing to the client %w", err)
				return
			}
		}
	}
}

func HandleConnection(conn *websocket.Conn) {

	// NOTE: (saad) we need to get the username for connection from the routes
	addConnection("test", conn)
	for {
		mty, b, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error reading from connection %w", err)
			break
		}
		fmt.Printf("recv  message %s\n", string(b))
		writeAll(mty, b)
		// if err := conn.WriteMessage(msgt,b); err != nil {
		// 	log.Printf("error reading from connection %w", err)
		// 	break
		// }
	}

}

// PASSWORD HASHING

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// JWT
var secretKey = []byte("secret-key")

func CreateToken(id uint) (string, error) {
	// sign it with user id 
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id": id,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
