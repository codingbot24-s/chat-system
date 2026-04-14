package db

import "gorm.io/gorm"

type User struct {
	gorm.Model	
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
	Messages []Message 
}

type Message struct {
	gorm.Model
	Content string 	`json:"content"`
	UserId  uint 	
}

