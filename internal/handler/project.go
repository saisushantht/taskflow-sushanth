package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"taskflow/internal/middleware"
	"taskflow/internal/model"
	"taskflow/internal/repository"
	"taskflow/internal/util"

	"github.com/go-chi/chi/v5"
)

type ProjectHandler struct {
	projectRepo *repository.ProjectRepository
	taskRepo    *repository.TaskRepository
}

func NewProjectHandler(projectRepo *repository.ProjectRepository, taskRepo *repository.TaskRepository) *ProjectHandler {
	return &ProjectHandler{projectRepo: projectRepo, taskRepo: taskRepo}
}

type projectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	projects, err := h.projectRepo.ListByUser(userID)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if projects == nil {
		projects = []model.Project{}
	}

	util.JSON(w, http.StatusOK, map[string]interface{}{"projects": projects})
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var req projectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		util.ValidationError(w, map[string]string{"name": "is required"})
		return
	}

	project, err := h.projectRepo.Create(req.Name, req.Description, userID)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	util.JSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	project, err := h.projectRepo.FindByID(id)
	if err == sql.ErrNoRows {
		util.ErrorResponse(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	tasks, _, err := h.taskRepo.ListByProject(id, "", "", 1, 100)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if tasks == nil {
		tasks = []model.Task{}
	}

	util.JSON(w, http.StatusOK, map[string]interface{}{
		"id":          project.ID,
		"name":        project.Name,
		"description": project.Description,
		"owner_id":    project.OwnerID,
		"created_at":  project.CreatedAt,
		"tasks":       tasks,
	})
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := r.Context().Value(middleware.UserIDKey).(string)

	project, err := h.projectRepo.FindByID(id)
	if err == sql.ErrNoRows {
		util.ErrorResponse(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if project.OwnerID != userID {
		util.ErrorResponse(w, http.StatusForbidden, "forbidden")
		return
	}

	var req projectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		req.Name = project.Name
	}
	if req.Description == "" {
		req.Description = project.Description
	}

	updated, err := h.projectRepo.Update(id, req.Name, req.Description)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	util.JSON(w, http.StatusOK, updated)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := r.Context().Value(middleware.UserIDKey).(string)

	project, err := h.projectRepo.FindByID(id)
	if err == sql.ErrNoRows {
		util.ErrorResponse(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if project.OwnerID != userID {
		util.ErrorResponse(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.projectRepo.Delete(id); err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) Stats(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := h.projectRepo.FindByID(id)
	if err == sql.ErrNoRows {
		util.ErrorResponse(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	stats, err := h.taskRepo.GetStats(id)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	util.JSON(w, http.StatusOK, stats)
}
