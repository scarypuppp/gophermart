package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

const (
	defaultTokenExpiresSeconds = 24 * 60 * 60 * 30 // 30 days
	defaultAccrualPollInterval = 2
)

type Config struct {
	SecretKey           string `env:"SECRET_KEY"`
	Address             string `env:"RUN_ADDRESS"`
	DatabaseURI         string `env:"DATABASE_URI"`
	AccrualSystemAddr   string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	AccrualPollInterval int    `env:"ACCRUAL_POLL_INTERVAL"`
	TokenExpSeconds     int64  `env:"TOKEN_EXP"`
}

func GetConfig() (*Config, error) {
	var config Config
	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	secretKeyFlag := flag.String("k", "", "application secret key")
	addrFlag := flag.String("a", "", "server address host:port")
	databaseUriFlag := flag.String("d", "", "database dsn string")
	accrualSystemAddrFlag := flag.String("r", "", "accrual system host:port")
	accrualPollIntervalFlag := flag.Int64("r", 0, "accrual system poll interval in seconds")
	tokenExpSecondsFlag := flag.Int64("e", 0, "jwt token expires in seconds")
	flag.Parse()

	if config.SecretKey == "" && *secretKeyFlag == "" {
		return nil, fmt.Errorf("secret key should not be empty")
	}
	if config.Address == "" && *addrFlag == "" {
		return nil, fmt.Errorf("service address should not be empty")
	}
	if config.DatabaseURI == "" && *databaseUriFlag == "" {
		return nil, fmt.Errorf("database uri should not be empty")
	}
	if config.AccrualSystemAddr == "" && *accrualSystemAddrFlag == "" {
		return nil, fmt.Errorf("accrual system address should not be empty")
	}
	if config.AccrualPollInterval == 0 && *accrualPollIntervalFlag == 0 {
		config.TokenExpSeconds = defaultAccrualPollInterval
	}
	if config.TokenExpSeconds == 0 && *tokenExpSecondsFlag == 0 {
		config.TokenExpSeconds = defaultTokenExpiresSeconds
	}

	return &config, nil
}
