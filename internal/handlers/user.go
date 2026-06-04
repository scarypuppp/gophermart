package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

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
func (h *Api) Register(w http.ResponseWriter, r *http.Request) {
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

func (h *Api) Login(w http.ResponseWriter, r *http.Request) {}
