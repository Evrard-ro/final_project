package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Evrard-ro/final_project/pkg/db"
)

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	// Десериализуем JSON
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем заголовок
	if strings.TrimSpace(task.Title) == "" {
		writeError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	// Проверяем дату
	if err := checkDate(&task); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Добавляем задачу в БД
	id, err := db.AddTask(&task)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем ID
	writeJSON(w, map[string]string{"id": strconv.FormatInt(id, 10)}, http.StatusOK)
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	writeJSON(w, task, http.StatusOK)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	// Десериализуем JSON
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем ID
	if task.ID == "" {
		writeError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	// Проверяем заголовок
	if strings.TrimSpace(task.Title) == "" {
		writeError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	// Проверяем дату
	if err := checkDate(&task); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Обновляем задачу в БД
	if err := db.UpdateTask(&task); err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}

	// Возвращаем пустой JSON
	writeJSON(w, map[string]string{}, http.StatusOK)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	if err := db.DeleteTask(id); err != nil {
		writeError(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]string{}, http.StatusOK)
}

func taskDoneHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		writeError(w, "Не указан идентификатор", http.StatusBadRequest)
		return
	}

	// Получаем задачу
	task, err := db.GetTask(id)
	if err != nil {
		writeError(w, "Задача не найдена", http.StatusNotFound)
		return
	}

	// Если правило повторения отсутствует - удаляем задачу
	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]string{}, http.StatusOK)
		return
	}

	// Парсим дату задачи
	taskDate, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Вычисляем следующую дату от даты задачи
	next, err := NextDate(taskDate, task.Date, task.Repeat)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Обновляем дату задачи
	if err := db.UpdateTaskDate(id, next); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{}, http.StatusOK)
}

func checkDate(task *db.Task) error {
	now := time.Now()

	// Если дата не указана, берём сегодняшнюю
	if task.Date == "" {
		task.Date = now.Format(DateFormat)
	}

	// Проверяем корректность даты
	t, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		return err
	}

	// Если есть правило повторения, проверяем его
	var next string
	if len(task.Repeat) > 0 {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return err
		}
	}

	// Если сегодня больше task.Date
	if afterNow(now, t) {
		if len(task.Repeat) == 0 {
			// Если правила нет, берём сегодняшнее число
			task.Date = now.Format(DateFormat)
		} else {
			// Иначе берём вычисленную следующую дату
			task.Date = next
		}
	}

	return nil
}

func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)

	response, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal server error"}`))
		return
	}
	w.Write(response)
}

func writeError(w http.ResponseWriter, errMsg string, statusCode int) {
	writeJSON(w, map[string]string{"error": errMsg}, statusCode)
}
