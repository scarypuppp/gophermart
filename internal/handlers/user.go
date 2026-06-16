package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/scarypuppp/gophermart/internal/auth"
	"github.com/scarypuppp/gophermart/internal/service"
	"go.uber.org/zap"
)

//
// REGISTER
//

type registerRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register godoc
//
//	@Summary		Регистрация пользователя
//	@Tags			auth
//	@Accept			json
//	@Produce		plain
//	@Param			body	body		registerRequest	true	"Данные пользователя"
//	@Success		200
//	@Failure		400	{string}	string	"Неверный формат запроса"
//	@Failure		409	{string}	string	"Логин уже занят"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/user/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var buffer bytes.Buffer
	var requestData registerRequest
	_, err := buffer.ReadFrom(r.Body)
	if err = json.Unmarshal(buffer.Bytes(), &requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.userService.RegisterUser(context.Background(), requestData.Login, requestData.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLoginAlreadyExists):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			h.logger.Error("Handler Register unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	token, err := auth.CreateToken(h.config.SecretKey, user.ID, h.config.TokenExpSeconds)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)
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

// Login godoc
//
//	@Summary		Аутентификация пользователя
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		loginRequest	true	"Данные пользователя"
//	@Success		200	{object}	loginResponse
//	@Failure		400	{string}	string	"Неверный формат запроса"
//	@Failure		401	{string}	string	"Неверный логин или пароль"
//	@Failure		500	{string}	string	"Внутренняя ошибка"
//	@Router			/api/user/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var buffer bytes.Buffer
	var requestData loginRequest
	_, err := buffer.ReadFrom(r.Body)
	if err = json.Unmarshal(buffer.Bytes(), &requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.userService.LoginUser(context.Background(), requestData.Login, requestData.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLoginPasswordNotExist):
			http.Error(w, err.Error(), http.StatusUnauthorized)
		default:
			h.logger.Error("Handler Login unhandled error", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	token, err := auth.CreateToken(h.config.SecretKey, user.ID, h.config.TokenExpSeconds)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	response, err := json.Marshal(loginResponse{UserID: user.ID, AccessToken: token})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
