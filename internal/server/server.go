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
	"github.com/scarypuppp/gophermart/internal/accrual_poller"
	"github.com/scarypuppp/gophermart/internal/config"
	"github.com/scarypuppp/gophermart/internal/handlers"
	"github.com/scarypuppp/gophermart/internal/repository"
	"github.com/scarypuppp/gophermart/internal/service"
	"go.uber.org/zap"

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
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Could not create logger: %v", err)
	}
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := repository.NewPostgresDB(s.Config.DatabaseURI)
	if err != nil {
		logger.Fatal("Could not create database connection", zap.Error(err))
	}

	if err := runMigrations(s.Config.DatabaseURI); err != nil {
		logger.Fatal("Error running migrations", zap.Error(err))
	}

	uow := repository.NewUnitOfWorkPostgres(db)
	userService := service.NewUserService(uow)
	orderService := service.NewOrderService(uow)
	transactionService := service.NewTransactionService(uow)
	handler := handlers.NewHandler(s.Config, logger, userService, orderService, transactionService)

	poller := accrual_poller.NewAccrualPoller(s.Config.AccrualSystemAddr, uow, logger)
	go poller.Run(ctx)

	srv := &http.Server{
		Addr:         s.Config.Address,
		Handler:      handler.GetRouter(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Could not start server", zap.Error(err))
			cancel()
		}
	}()
	logger.Info("Server started", zap.String("address", s.Config.Address))
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	cancel()

	shutdownCtx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()
	logger.Info("Server stopped gracefully")
	return srv.Shutdown(shutdownCtx)
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
