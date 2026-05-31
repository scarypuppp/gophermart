package server

import (
	"net/http"
	"time"

	"github.com/scarypuppp/gophermart/internal/config"
)

type Server struct {
	Addr   string
	Server *http.Server
}

func NewServer(addr string, server *http.Server) Server {
	return Server{Addr: addr, Server: server}
}

func (s *Server) Run() error {
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}
	srv := &http.Server{
		Addr:         cfg,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return nil
}
