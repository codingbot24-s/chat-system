package db

import "gorm.io/gorm"



/*
	NOTE : we can save all the messages thats been recvied 
	by this user currently we are saving the one thats been 
	sended
*/ 
type User struct {
	gorm.Model
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Messages []Message `gorm:"foreignKey:UserId"`
}

type Message struct {
	gorm.Model
	Content []byte
	UserId  uint
	User    User `gorm:"constraint:OnDelete:CASCADE"`
}
