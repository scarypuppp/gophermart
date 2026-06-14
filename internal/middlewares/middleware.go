package middlewares

import (
	"github.com/scarypuppp/gophermart/internal/config"
	"go.uber.org/zap"
)

type Middleware struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewMiddleware(cfg *config.Config, logger *zap.Logger) Middleware {
	return Middleware{cfg, logger}
}
