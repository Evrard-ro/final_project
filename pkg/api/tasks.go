package api

import (
	"net/http"

	"github.com/Evrard-ro/final_project/pkg/db"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.FormValue("search")
	
	tasks, err := db.Tasks(50, search)
	if err != nil {
		writeError(w, err.Error())
		return
	}
	writeJSON(w, TasksResp{
		Tasks: tasks,
	})
}