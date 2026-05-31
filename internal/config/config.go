package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	SecretKey         string `env:"SECRET_KEY"`
	Address           string `env:"RUN_ADDRESS"`
	DatabaseURI       string `env:"DATABASE_URI"`
	AccrualSystemAddr string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func GetConfig() (*Config, error) {
	var config Config
	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	secretKeyFlag := flag.String("k", "", "application secret key")
	addrFlag := flag.String("a", "", "server address host:port")
	databaseUriFlag := flag.String("d", "", "database dsn string")
	accrualSystemAddr := flag.String("r", "", "accrual system host:port")
	flag.Parse()

	if config.SecretKey == "" && *secretKeyFlag == "" {
		return nil, fmt.Errorf("secret key should not be empty")
	}
	if config.Address == "" && *addrFlag == "" {
		return nil, fmt.Errorf("secret key should not be empty")
	}
	if config.DatabaseURI == "" && *databaseUriFlag == "" {
		return nil, fmt.Errorf("secret key should not be empty")
	}
	if config.AccrualSystemAddr == "" && *accrualSystemAddr == "" {
		return nil, fmt.Errorf("secret key should not be empty")
	}

	return &config, nil
}
