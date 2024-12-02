package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	logger "github.com/thalq/gopher_mart/internal/middleware"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthHandler struct {
	service *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (req *AuthRequest) Validate() error {
	if req.Login == "" {
		return fmt.Errorf("login is empty")
	}
	if req.Password == "" {
		return fmt.Errorf("password is empty")
	}
	return nil
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать тело запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	logger.Sugar.Infof("Got request: %s", string(body))
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Не удалось распарсить JSON", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if userExists, err := h.service.CheckUserExists(req.Login); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if userExists {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	userID, err := h.service.Register(req.Login, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Sugar.Infof("User %s registered", req.Login)

	if err := h.service.CreateUserBalance(userID); err != nil {
		http.Error(w, "Failed to create user balance account", http.StatusInternalServerError)
		return
	}
	logger.Sugar.Infof("User balance account created for user %s", req.Login)

	token := h.service.GenerateToken(userID)
	http.SetCookie(w, &http.Cookie{
		Name:    "Authorization",
		Value:   token,
		Expires: time.Now().Add(time.Hour * 24),
		Path:    "/",
	})
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать тело запроса", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	logger.Sugar.Infof("Got request: %s", string(body))
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Не удалось распарсить JSON", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	autheticated, userID, err := h.service.Authenticate(req.Login, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !autheticated {
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}
	logger.Sugar.Infof("User %s authenticated", req.Login)

	token := h.service.GenerateToken(userID)
	http.SetCookie(w, &http.Cookie{
		Name:    "Authorization",
		Value:   token,
		Expires: time.Now().Add(time.Hour * 24),
		Path:    "/",
	})

	w.WriteHeader(http.StatusOK)
}
