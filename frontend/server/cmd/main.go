package main

import (
	"example_fe/base"
	"example_fe/base/db"
	"log"
)

func main() {
	dbConn, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer dbConn.Close()

	data := &db.DataStore{
		Auth: &db.DefaultAuthStore{DBConn: dbConn},
	}

	base.Server(data)
}
