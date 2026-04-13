package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"taskflow/internal/repository"
	"taskflow/internal/util"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtSecret string) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, jwtSecret: jwtSecret}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	fields := map[string]string{}
	if req.Name == "" {
		fields["name"] = "is required"
	}
	if req.Email == "" {
		fields["email"] = "is required"
	}
	if req.Password == "" {
		fields["password"] = "is required"
	}
	if len(req.Password) < 6 {
		fields["password"] = "must be at least 6 characters"
	}
	if len(fields) > 0 {
		util.ValidationError(w, fields)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	user, err := h.userRepo.Create(req.Name, req.Email, string(hashed))
	if err != nil {
		util.ErrorResponse(w, http.StatusConflict, "email already exists")
		return
	}

	token, err := h.generateToken(user.ID, user.Email)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	util.JSON(w, http.StatusCreated, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	fields := map[string]string{}
	if req.Email == "" {
		fields["email"] = "is required"
	}
	if req.Password == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		util.ValidationError(w, fields)
		return
	}

	user, err := h.userRepo.FindByEmail(req.Email)
	if err == sql.ErrNoRows {
		util.ErrorResponse(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		util.ErrorResponse(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := h.generateToken(user.ID, user.Email)
	if err != nil {
		util.ErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	util.JSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

func (h *AuthHandler) generateToken(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
