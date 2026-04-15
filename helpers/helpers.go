package helpers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/codingbot24.s/chat-system/db"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Clients []map[uint]*websocket.Conn

var (
	Connection Clients
	mu         sync.Mutex
)

func addConnection(userID uint, conn *websocket.Conn) {

	mu.Lock()
	Connection = append(Connection, map[uint]*websocket.Conn{userID: conn})
	defer mu.Unlock()
}

 
func CreateRecvMessage(DBh *gorm.DB,b []byte,sender string,userID uint) {
	// find user by this id in db and 
	// create a recv message for this user
	var user db.User
	DBh.First(&user, 10)
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

func CreateMessage(userID uint, b []byte) *db.Message {
	m := &db.Message{
		Content: b,
		UserId:  userID,
	}

	return m
}

// FIND USER will find first user with given id 
func findUser(userId uint, DBh *gorm.DB) (*db.User, error) {
	var user db.User
	if result := DBh.First(&user, userId); result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// STORE Message In DB
func StoreMessage(b []byte, DBh *gorm.DB, userID uint) error {
	msg := CreateMessage(userID, b)
	// find the user with this id
	user, err := findUser(userID, DBh)
	if err != nil {
		return err
	}

	// create a message in this user
	if err := DBh.Model(user).Association("Messages").Append(&db.Message{Content: msg.Content, UserId: user.ID}); err != nil {
		return err 
	}	
	return nil
}

func writeAllMessages(userID uint,DBh *gorm.DB, conn *websocket.Conn) error {
	user,err  := findUser(userID,DBh)
	if err != nil {
		return err
	}
	var msgs []db.Message
	if err := DBh.Model(&user).Association("Messages").Find(&msgs);err != nil {
		return err
	}

	// write this content back to user
	for i := 0; i < len(msgs); i++ {
		if err := conn.WriteMessage(websocket.TextMessage,msgs[i].Content); err != nil {
			return err
		}
	}
	return nil 
}

func HandleConnection(userID uint, conn *websocket.Conn, db *gorm.DB) {

	addConnection(userID, conn)
	for {
		// we can save this message in db in byte
		writeAllMessages(userID,db,conn)
		mty, b, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error reading from connection %w", err)
			break
		}
		fmt.Printf("recv  message %s\n", string(b))
		writeAll(mty, b)

		if err := StoreMessage(b, db, userID); err != nil {
			log.Printf("error storing message: %w", err)
		}
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



/*
	AUTH MIDDLEWARE
*/


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

		id, ok := claims["id"]
		if !ok {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserId(ctx context.Context) float64 {
	userId := ctx.Value("id").(float64)
	return userId
}
