package api

import (
	"net/http"

	"github.com/Evrard-ro/final_project/pkg/db"
)

const (
	DefaultTasksLimit = 50
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.FormValue("search")

	tasks, err := db.Tasks(DefaultTasksLimit, search)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, TasksResp{Tasks: tasks}, http.StatusOK)
}
