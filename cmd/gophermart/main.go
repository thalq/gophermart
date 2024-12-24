package main

import (
	"net/http"
	// "os/exec"

	logger "github.com/thalq/gopher_mart/internal/middleware"
	"github.com/thalq/gopher_mart/pkg/config"
	router "github.com/thalq/gopher_mart/pkg/http"
	"github.com/thalq/gopher_mart/pkg/storage"
)

func main() {
	logger.InitLogger()
	cfg := config.NewConfig()

	storage.InitDB(cfg.DatabaseURI)
	router := router.NewRouter(cfg)

	// cmd := exec.Command("./accrual_darwin_arm64", cfg.AccrualSystemAddress)
	// output, err := cmd.CombinedOutput()

	// if err != nil {
	// 	logger.Sugar.Fatalf("Ошибка запуска исполняемого файла: %s", err)
	// 	return
	// }
	// logger.Sugar.Infof("Исполняемы файл запущен %s", string(output))

	logger.Sugar.Infof("Starting server on %s", cfg.RunAdress)
	if err := http.ListenAndServe(cfg.RunAdress, router); err != nil {
		logger.Sugar.Fatalf("Error run server: %s", err)
	}
}
