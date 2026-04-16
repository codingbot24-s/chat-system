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

type Clients map[uint]*websocket.Conn

var (
	Connection = make(Clients)
	mu         sync.Mutex
)

func addConnection(userID uint, conn *websocket.Conn) {

	mu.Lock()
	Connection[userID] = conn
	mu.Unlock()
}

func removeConnection(userID uint) {
	mu.Lock()
	delete(Connection, userID)
	mu.Unlock()
}

func CreateRecvMessage(sender string,msg []byte,receiverId uint) *db.RecvMessage {
	var rcvMsg db.RecvMessage
	rcvMsg.SendedBy = sender
	rcvMsg.Content = msg
	rcvMsg.UserId = receiverId

	return &rcvMsg
}


func StoreRecvMessage(db *gorm.DB, msgToStore *db.RecvMessage) error {	
	receiver,err  := findUser(msgToStore.UserId,db)
	if err != nil {
		return err 
	}

	if err := db.Model(receiver).Association("RecvMessages").Append(msgToStore); err != nil {
		return err

	}

	return nil 
}


func writeAll(ty int, msg []byte, senderName string,DBh *gorm.DB) {
	mu.Lock()
	for receiverId, conn := range Connection {
		if err := conn.WriteMessage(ty, msg); err != nil {
			log.Printf("error writing to the client %w", err)
			delete(Connection, receiverId)
		}
		rcvMsg := CreateRecvMessage(senderName,msg,receiverId)
		StoreRecvMessage(DBh,rcvMsg)
	}
	defer mu.Unlock()
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
	// create a sended messages
	msg := CreateMessage(userID, b)
	// find the user with this id
	user, err := findUser(userID, DBh)
	if err != nil {
		return err
	}

	// create a sended message in this user
	if err := DBh.Model(user).Association("Messages").Append(&db.Message{Content: msg.Content, UserId: user.ID}); err != nil {
		return err
	}


	return nil
}

// send all the messages back when user connected
func SendBack(userId uint,DBh *gorm.DB, conn *websocket.Conn ) error {	
	var allMessages []struct {
		Content  []byte
		Sender   string
		CreatedAt time.Time
	}

	// Fetch all sent messages with sender info
	if err := DBh.Table("messages").Select("messages.content, users.name as sender, messages.created_at").
		Joins("JOIN users ON users.id = messages.user_id").
		Order("messages.created_at ASC").
		Find(&allMessages).Error; err != nil {
		return err
	}

	// Send all messages in chronological order
	for _, msg := range allMessages {
		messageWithSender := fmt.Sprintf("[%s] %s: %s", msg.CreatedAt.Format("15:04:05"), msg.Sender, string(msg.Content))
		if err := conn.WriteMessage(websocket.TextMessage, []byte(messageWithSender)); err != nil {
			return err
		}
	}

	return nil 
}



func HandleConnection(userID uint, conn *websocket.Conn, db *gorm.DB) {

	addConnection(userID, conn)
	sender,err := findUser(userID,db)	
	if err != nil {
		log.Println("cant get the sender")
		return
	}
	SendBack(userID,db,conn)
	for {
		mty, b, err := conn.ReadMessage()
		if err != nil {
			log.Printf("error reading from connection %w", err)
			break
		}
		fmt.Printf("recv  message %s\n", string(b))
		writeAll(mty,b,sender.Name,db)
		

		// Storing the messages that this user is sending 
		if err := StoreMessage(b, db, userID); err != nil {
			log.Printf("error storing message: %w", err)
		}
		
	}

	removeConnection(userID)
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
