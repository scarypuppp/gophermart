package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/scarypuppp/gophermart/internal/auth"
	"github.com/scarypuppp/gophermart/internal/service"
)

//
// REGISTER
//

type registerRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register Хэндлер для регистрации пользователя
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var buffer bytes.Buffer
	var requestData registerRequest
	_, err := buffer.ReadFrom(r.Body)
	if err = json.Unmarshal(buffer.Bytes(), &requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = h.userService.RegisterUser(context.Background(), requestData.Login, requestData.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLoginAlreadyExists):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
	w.WriteHeader(http.StatusOK)
}

// LOGIN
type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	UserID      int64  `json:"user_id"`
	AccessToken string `json:"access_token"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var buffer bytes.Buffer
	var requestData registerRequest
	_, err := buffer.ReadFrom(r.Body)
	if err = json.Unmarshal(buffer.Bytes(), &requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.userService.LoginUser(context.Background(), requestData.Login, requestData.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLoginPasswordNotExist):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	token, err := auth.CreateToken(h.config.SecretKey, user.ID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	response, err := json.Marshal(loginResponse{UserID: user.ID, AccessToken: token})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
