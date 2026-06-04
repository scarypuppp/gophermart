package main

import (
	"log"

	"github.com/scarypuppp/gophermart/internal/config"
	"github.com/scarypuppp/gophermart/internal/server"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		panic(err)
	}
	s := server.NewServer(cfg)
	if err := s.Run(); err != nil {
		log.Fatalf("Error running server")
	}
}
