package handlers

import (
	"encoding/json"
	"net/http"
	"go-realtime-workspace/models"
	"go-realtime-workspace/repository"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// TaskHandler handles task-related HTTP requests.
type TaskHandler struct {
	repo *repository.TaskRepository
}

// NewTaskHandler creates a new task handler.
func NewTaskHandler(repo *repository.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

// Create handles task creation.
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userId"]

	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Missing required field: title", http.StatusBadRequest)
		return
	}

	// Set default priority if not provided
	if req.Priority == "" {
		req.Priority = models.TaskPriorityMedium
	}

	task, err := h.repo.Create(r.Context(), userID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// GetByID handles retrieving a task by ID.
func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	task, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// GetByUser handles retrieving all tasks for a user.
func (h *TaskHandler) GetByUser(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userId"]
	status := r.URL.Query().Get("status")

	tasks, err := h.repo.GetByUserID(r.Context(), userID, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetDueSoon handles retrieving tasks that are due soon.
func (h *TaskHandler) GetDueSoon(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userId"]

	// Default to 24 hours
	hoursStr := r.URL.Query().Get("hours")
	hours := 24
	if hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil {
			hours = h
		}
	}

	tasks, err := h.repo.GetDueSoon(r.Context(), userID, time.Duration(hours)*time.Hour)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// Update handles task updates.
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task, err := h.repo.Update(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// Delete handles task deletion.
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := h.repo.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
