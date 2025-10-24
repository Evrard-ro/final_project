package api

import "net/http"

func Init() {

	InitAuth()

	http.HandleFunc("/api/nextdate", nextDayHandler)
	http.HandleFunc("/api/task", auth(taskHandler))
	http.HandleFunc("/api/tasks", auth(tasksHandler))
	http.HandleFunc("/api/task/done", auth(taskDoneHandler))
	http.HandleFunc("/api/signin", signInHandler)
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		deleteTaskHandler(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}
