package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Evrard-ro/final_project/pkg/api"
)


func getPort() string {
	// Проверяем переменную окружения TODO_PORT
	if port := os.Getenv("TODO_PORT"); port != "" {
		return port
	}
	// По умолчанию используем порт 7540
	return "7540"
}

func getWebDir() string {
	// Получаем текущую рабочую директорию
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(wd, "web")
}

func Run() {
	port := getPort()
	webDir := getWebDir()
	api.Init() // ← добавьте эту строку
	fs := http.FileServer(http.Dir(webDir))
	http.Handle("/", fs)
	log.Printf("Сервер запущен на http://localhost:%s/", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
