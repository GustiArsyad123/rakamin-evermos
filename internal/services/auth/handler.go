package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/example/ms-ecommerce/internal/pkg/models"
	"github.com/gorilla/mux"
)

type Handler struct {
	usecase Usecase
}

func NewHandler(usecase Usecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
	}

	token, err := h.usecase.Register(user)
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	token, err := h.usecase.Login(req.Email, req.Password)
	if err != nil {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is missing in parameters"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}

	user, err := h.usecase.GetUserByID(id)
	if err != nil {
		log.Printf("Error getting user by ID %d: %v", id, err)
		h.writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	h.writeJSON(w, http.StatusOK, user)
}

func (h *Handler) handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.usecase.GetAllUsers()
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		http.Error(w, "failed to retrieve users", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": users,
	})
}

func (h *Handler) handleGetMyProfile(w http.ResponseWriter, r *http.Request) {
	// The user ID is injected into the context by the JWTAuth middleware.
	userID, ok := middleware.GetUserID(r)
	if !ok {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "could not identify user"})
		return
	}

	user, err := h.usecase.GetUserByID(userID)
	if err != nil {
		log.Printf("Error getting user profile for ID %d: %v", userID, err)
		h.writeJSON(w, http.StatusNotFound, map[string]string{"error": "user profile not found"})
		return
	}

	h.writeJSON(w, http.StatusOK, user)
}

func (h *Handler) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	resetToken, err := h.usecase.ForgotPassword(req.Email)
	if err != nil {
		// To prevent email enumeration, we can return a success response even if the user is not found.
		// However, for development, returning the error is fine.
		h.writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"reset_token": resetToken})
}

func (h *Handler) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.usecase.ResetPassword(req.Token, req.Password); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleSSOLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	token, err := h.usecase.SSOLogin(req.IDToken)
	if err != nil {
		h.writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
