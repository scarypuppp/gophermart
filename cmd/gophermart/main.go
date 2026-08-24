// Package main is the entry point of the gophermart service.
//
// @title           Gophermart API
// @version         1.0
// @description     Loyalty system API
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
//
// @host      localhost:8080
// @BasePath  /
package main

import (
	"log"

	_ "github.com/scarypuppp/gophermart/docs"
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
