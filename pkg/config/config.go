package config

import (
	"flag"
	"os"

	logger "github.com/thalq/gopher_mart/internal/middleware"
)

type Config struct {
	RunAdress            string `env:"RUN_ADDRESS" json:"run_address"`
	DatabaseURI          string `env:"DATABASE_URI" json:"database_uri"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" json:"accrual_system_address"`
}

func getEnv(value string, defaultValue string) string {
	if value, exist := os.LookupEnv(value); exist {
		return value
	}
	return defaultValue
}

func NewConfig() *Config {
	defaultRunAdress := "localhost:8081"
	defaultDatabaseURI := "postgres://postgres:postgres@localhost/postgres?sslmode=disable"
	envRunAddress := getEnv("RUN_ADDRESS", defaultRunAdress)
	envDatabaseURI := getEnv("DATABASE_URI", defaultDatabaseURI)
	envAccrualSystemAddress := getEnv("ACCRUAL_SYSTEM_ADDRESS", "http://localhost:8080")

	runAddress := flag.String("a", envRunAddress, "address to run server")
	databaseURI := flag.String("d", envDatabaseURI, "database URI")
	accrualSystemAddress := flag.String("r", envAccrualSystemAddress, "accrual system address")

	logger.Sugar.Infof("Run address: %v, Database URI: %v, Accrual system address: %v", runAddress, databaseURI, accrualSystemAddress)

	flag.Parse()

	return &Config{
		RunAdress:            *runAddress,
		DatabaseURI:          *databaseURI,
		AccrualSystemAddress: *accrualSystemAddress,
	}
}
