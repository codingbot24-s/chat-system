package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/codingbot24.s/chat-system/db"
	"github.com/codingbot24.s/chat-system/helpers"
	rqrstype "github.com/codingbot24.s/chat-system/types"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) handler {
	return handler{db}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (h *handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (h *handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error upgarding con %w", err)
		return
	}
	go helpers.HandleConnection(conn)
}

// we need better error handling

func (h *handler) SignUp(w http.ResponseWriter, r *http.Request)  {
	var body rqrstype.SignUpBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 
	}

	var usr db.User
	if result := h.db.Where("name = ?", body.Name).First(&usr); result.
	Error != gorm.ErrRecordNotFound {
		http.Error(w, "user already exists", http.StatusBadRequest)
		return 
	}

	usr.Name = body.Name
	usr.Email = body.Email
	hashedPass, err := helpers.HashPassword(body.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	usr.Password = hashedPass
	result := h.db.Create(&usr)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusBadRequest)
		return 
	}

	token,err := helpers.CreateToken(usr.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 
	}

	
	var res rqrstype.SignUpres 
	res.Success = true
	res.Msg = fmt.Sprintf("user with name created successfully %s",usr.Name)
	res.Token = token

	json.NewEncoder(w).Encode(res)

}


func(h *handler) Login(w http.ResponseWriter, r *http.Request) {
	var body rqrstype.LoginReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 
	}

    var usr db.User
	if result := h.db.Where("email = ?", body.Email).First(&usr);
	result.Error == gorm.ErrRecordNotFound {
		http.Error(w, "user with email not found", http.StatusBadRequest)
		return
	}

	if !helpers.VerifyPassword(body.Password, usr.Password) {
		http.Error(w, "password is incorrect", http.StatusBadRequest)
		return
	}

	token,err := helpers.CreateToken(usr.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return 
	}

	var res rqrstype.LoginRes 
	res.Msg = fmt.Sprintf("user with name %s logged in successfully ",usr.Name)
	res.Token = token

	json.NewEncoder(w).Encode(res)
}