package db

import "gorm.io/gorm"




/*
	TODO: when user connect it should recv all the meesage that he sended or recived 
*/

type User struct {
	gorm.Model
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Messages []Message `gorm:"foreignKey:UserId"`
	RecvMessages []RecvMessage `gorm:"foreignKey:UserId"`
}

type Message struct {
	gorm.Model
	Content []byte
	UserId  uint
	User    User `gorm:"constraint:OnDelete:CASCADE"`
}


type RecvMessage struct {
	gorm.Model
	SendedBy string
	Content []byte
	UserId  uint
	User    User `gorm:"constraint:OnDelete:CASCADE"`
}