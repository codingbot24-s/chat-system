package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


func Init () *gorm.DB {
	dsn := "postgresql://postgres:mysecretpassword@localhost:5432/postgres"	

	conn,err := gorm.Open(postgres.Open(dsn),&gorm.
	Config{})
	if err != nil {
		log.Fatalf("error connecting the db %w",err);
		return nil
	}
	conn.AutoMigrate(User{},Message{});
	return conn;
}