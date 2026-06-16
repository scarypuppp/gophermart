package config

import (
	"flag"
	"fmt"
	"os"

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
	return parseConfig(os.Args[1:])
}

func parseConfig(args []string) (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	fs := flag.NewFlagSet("gophermart", flag.ContinueOnError)
	secretKeyFlag := fs.String("k", "", "application secret key")
	addrFlag := fs.String("a", "", "server address host:port")
	databaseURIFlag := fs.String("d", "", "database dsn string")
	accrualAddrFlag := fs.String("r", "", "accrual system host:port")
	accrualIntervalFlag := fs.Int64("i", 0, "accrual system poll interval in seconds")
	tokenExpFlag := fs.Int64("e", 0, "jwt token expires in seconds")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if cfg.SecretKey == "" {
		if *secretKeyFlag == "" {
			return nil, fmt.Errorf("secret key should not be empty")
		}
		cfg.SecretKey = *secretKeyFlag
	}
	if cfg.Address == "" {
		if *addrFlag == "" {
			return nil, fmt.Errorf("service address should not be empty")
		}
		cfg.Address = *addrFlag
	}
	if cfg.DatabaseURI == "" {
		if *databaseURIFlag == "" {
			return nil, fmt.Errorf("database uri should not be empty")
		}
		cfg.DatabaseURI = *databaseURIFlag
	}
	if cfg.AccrualSystemAddr == "" {
		if *accrualAddrFlag == "" {
			return nil, fmt.Errorf("accrual system address should not be empty")
		}
		cfg.AccrualSystemAddr = *accrualAddrFlag
	}
	if cfg.AccrualPollInterval == 0 {
		if *accrualIntervalFlag != 0 {
			cfg.AccrualPollInterval = int(*accrualIntervalFlag)
		} else {
			cfg.AccrualPollInterval = defaultAccrualPollInterval
		}
	}
	if cfg.TokenExpSeconds == 0 {
		if *tokenExpFlag != 0 {
			cfg.TokenExpSeconds = *tokenExpFlag
		} else {
			cfg.TokenExpSeconds = defaultTokenExpiresSeconds
		}
	}

	return &cfg, nil
}
