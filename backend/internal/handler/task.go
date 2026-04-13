package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"taskflow/internal/middleware"
	"taskflow/internal/model"
	"taskflow/internal/repository"
	"taskflow/internal/util"
	"time"

	"github.com/go-chi/chi/v5"
)

type TaskHandler struct {
	taskRepo    *repository.TaskRepository
	projectRepo *repository.ProjectRepository
}

func NewTaskHandler(taskRepo *repository.TaskRepository, projectRepo *repository.ProjectRepository) *TaskHandler {
	return &TaskHandler{taskRepo: taskRepo, projectRepo: projectRepo}
}

type taskRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	status := r.URL.Query().Get("status")
	assigneeID := r.URL.Query().Get("assignee")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	tasks, total, err := h.taskRepo.ListByProject(projectID, status, assigneeID, page, limit)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if tasks == nil {
		tasks = []model.Task{}
	}

	util.JSON(w, http.StatusOK, map[string]interface{}{
		"tasks": tasks,
		"pagination": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	_, err := h.projectRepo.FindByID(projectID)
	if err == sql.ErrNoRows {
		util.ErrorResponse(w, http.StatusNotFound, "not found")
		return
	}

	var req taskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		util.ValidationError(w, map[string]string{"title": "is required"})
		return
	}
	if req.Status == "" {
		req.Status = "todo"
	}
	if req.Priority == "" {
		req.Priority = "medium"
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			util.ValidationError(w, map[string]string{"due_date": "must be in YYYY-MM-DD format"})
			return
		}
		dueDate = &parsed
	}

	task, err := h.taskRepo.Create(req.Title, req.Description, req.Status, req.Priority, projectID, req.AssigneeID, dueDate)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	util.JSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	task, err := h.taskRepo.FindByID(id)
	if err == sql.ErrNoRows {
		util.ErrorResponse(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var req taskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		req.Title = task.Title
	}
	if req.Description == "" {
		req.Description = task.Description
	}
	if req.Status == "" {
		req.Status = task.Status
	}
	if req.Priority == "" {
		req.Priority = task.Priority
	}
	if req.AssigneeID == nil {
		req.AssigneeID = task.AssigneeID
	}

	var dueDate *time.Time
	if req.DueDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			util.ValidationError(w, map[string]string{"due_date": "must be in YYYY-MM-DD format"})
			return
		}
		dueDate = &parsed
	} else {
		dueDate = task.DueDate
	}

	updated, err := h.taskRepo.Update(id, req.Title, req.Description, req.Status, req.Priority, req.AssigneeID, dueDate)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	util.JSON(w, http.StatusOK, updated)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := r.Context().Value(middleware.UserIDKey).(string)

	task, err := h.taskRepo.FindByID(id)
	if err == sql.ErrNoRows {
		util.ErrorResponse(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	project, err := h.projectRepo.FindByID(task.ProjectID)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if project.OwnerID != userID {
		util.ErrorResponse(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.taskRepo.Delete(id); err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
