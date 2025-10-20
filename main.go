package main

import (
	"log"
	"os"

	"github.com/Evrard-ro/final_project/pkg/db"
	"github.com/Evrard-ro/final_project/pkg/server"
)

func main() {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}
	if err := db.Init(dbFile); err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}

	server.Run()
}
