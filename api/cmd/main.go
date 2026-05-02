package main

import (
	"log"

	"example_api/base"
	"example_api/base/db"
	"example_api/config"
)

// @title           Example API
// @version         1.0
// @description     API backend for "Example App"
// @host            localhost:8080
// @BasePath        /
func main() {
	if !config.EnsureValidConfig() {
		log.Fatalln("Config is invalid")
	}

	dbConn, err := db.InitDB()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer dbConn.Close()

	go db.CleanupTasks(dbConn)

	data := &db.DataStore{
		Auth: &db.DefaultAuthStore{DBConn: dbConn},
	}

	srv := base.NewServer(data)
	srv.Run(config.LISTENER)
}
