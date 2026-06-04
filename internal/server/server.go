package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/scarypuppp/gophermart/internal/config"
	"github.com/scarypuppp/gophermart/internal/handlers"
	"github.com/scarypuppp/gophermart/internal/repository"
	"github.com/scarypuppp/gophermart/internal/service"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Server struct {
	Config *config.Config
}

func NewServer(config *config.Config) *Server {
	return &Server{config}
}

func (s *Server) Run() error {
	db, err := repository.NewPostgresDB(s.Config.DatabaseURI)
	if err != nil {
		log.Fatalf("Could not create database connection: %v", err)
	}

	if err := runMigrations(s.Config.DatabaseURI); err != nil {
		log.Fatal(err)
	}

	uow := repository.NewUnitOfWorkPostgres(db)
	us := service.NewUserService(uow)
	api := handlers.NewApi(us)

	srv := &http.Server{
		Addr:         s.Config.Address,
		Handler:      api.GetRouter(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return srv.Shutdown(ctx)
}

func runMigrations(dsn string) error {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return fmt.Errorf("create migrate: %w", err)
	}
	defer m.Close()

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
