package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

const DATE_FORMAT = "20060102"

type Handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return Handler{service: service}
}

func (h Handler) handleTask(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getTask(w, r)
	case http.MethodPost:
		h.postTask(w, r)
	case http.MethodPut:
		h.putTask(w, r)
	case http.MethodDelete:
		h.deleteTask(w, r)
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func (h Handler) getNextDate(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	now, err := time.Parse(DATE_FORMAT, nowStr)
	if err != nil {
		http.Error(w, "неверный формат даты", http.StatusBadRequest)
		return
	}

	next, err := h.service.NextDate(now, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.writeResponse(w, http.StatusOK, []byte(next))
}

func (h Handler) postTask(w http.ResponseWriter, r *http.Request) {
	var newTask Task

	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if newTask.Title == "" {
		http.Error(w, wrappError("Не заданы обязательные параметры"), http.StatusBadRequest)
		return
	}

	if newTask.Date == "" {
		newTask.Date = time.Now().Truncate(24 * time.Hour).Format(DATE_FORMAT)
	}

	date, err := time.Parse(DATE_FORMAT, newTask.Date)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	var nextDate string
	if newTask.Repeat == "" {
		nextDate = time.Now().Format(DATE_FORMAT)
	} else {
		nextDate, err = h.service.NextDate(time.Now().Truncate(24*time.Hour), newTask.Date, newTask.Repeat)
		if err != nil {
			http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
			return
		}
	}

	if date.Before(time.Now().Truncate(24 * time.Hour)) {
		newTask.Date = nextDate
	}

	createdTask, err := h.service.Create(newTask)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(map[string]interface{}{"id": createdTask.Id})
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	h.writeResponse(w, http.StatusCreated, resp)
}

func (h Handler) getTasks(w http.ResponseWriter, r *http.Request) {
	search := r.FormValue("search")
	tasks, err := h.service.FindBy(search)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	if tasks == nil {
		tasks = make([]Task, 0)
	}

	resp, err := json.Marshal(map[string]interface{}{"tasks": tasks})
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	h.writeResponse(w, http.StatusOK, resp)
}

func (h Handler) getTask(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		http.Error(w, wrappError("Не указан идентификатор"), http.StatusBadRequest)
		return
	}

	n, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	task, err := h.service.FindById(n)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(task)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	h.writeResponse(w, http.StatusOK, resp)
}

func (h Handler) putTask(w http.ResponseWriter, r *http.Request) {
	var task Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if task.Id == "" || task.Title == "" {
		http.Error(w, wrappError("Не указаны необходимые параметры"), http.StatusBadRequest)
		return
	}

	if _, err := strconv.Atoi(task.Id); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	if _, err := time.Parse(DATE_FORMAT, task.Date); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	if task.Repeat != "" {
		if _, err := h.service.NextDate(time.Now(), task.Date, task.Repeat); err != nil {
			http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
			return
		}
	}

	if err := h.service.Update(task); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	h.writeResponse(w, http.StatusAccepted, []byte("{}"))
}

func (h Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
	number := r.FormValue("id")
	id, err := strconv.Atoi(number)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id); err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
		return
	}

	h.writeResponse(w, http.StatusAccepted, []byte("{}"))
}

func (h Handler) postDone(w http.ResponseWriter, r *http.Request) {
	number := r.FormValue("id")
	id, err := strconv.Atoi(number)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	task, err := h.service.FindById(id)
	if err != nil {
		http.Error(w, wrappError(err.Error()), http.StatusBadRequest)
		return
	}

	if task.Repeat == "" {
		if err := h.service.Delete(id); err != nil {
			http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
			return
		}
	} else {
		nextDate, err := h.service.NextDate(time.Now().Truncate(24*time.Hour), task.Date, task.Repeat)
		if err != nil {
			http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
			return
		}

		task.Date = nextDate
		if err := h.service.Update(task); err != nil {
			http.Error(w, wrappError(err.Error()), http.StatusInternalServerError)
			return
		}
	}

	h.writeResponse(w, http.StatusOK, []byte("{}"))
}

func (h Handler) writeResponse(w http.ResponseWriter, statusCode int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(body); err != nil {
		log.Println(err.Error())
	}
}

func wrappError(message string) string {
	str, _ := json.Marshal(map[string]interface{}{"error": message})
	return string(str)
}

func (h Handler) InitRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Handle("/*", http.FileServer(http.Dir("./web")))
	r.Get("/api/nextdate", h.getNextDate)
	r.Get("/api/tasks", h.getTasks)
	r.MethodFunc("GET", "/api/task", h.handleTask)
	r.MethodFunc("POST", "/api/task", h.handleTask)
	r.MethodFunc("PUT", "/api/task", h.handleTask)
	r.MethodFunc("DELETE", "/api/task", h.handleTask)
	r.Post("/api/task/done", h.postDone)
	return r
}
