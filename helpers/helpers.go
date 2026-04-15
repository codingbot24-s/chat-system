package helpers

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

/*
	Password Hashing
*/

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

/*
	JWT
*/

var secretKey = []byte("secret-key")

func CreateToken(id uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":  float64(id),
			"exp": time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "not authorized", http.StatusUnauthorized)
			return
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secretKey, nil
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		id,ok := claims["id"]
		if !ok {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(),"id",id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
 

func GetUserId(ctx context.Context) float64 {
	userId := ctx.Value("id").(float64)
	return userId
}