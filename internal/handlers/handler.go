package handlers

import "github.com/scarypuppp/gophermart/internal/service"

type Api struct {
	userService *service.UserService
}

func NewApi(us *service.UserService) *Api {
	return &Api{
		userService: us,
	}
}
