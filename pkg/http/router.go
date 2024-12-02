package http

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thalq/gopher_mart/internal/auth"
	"github.com/thalq/gopher_mart/internal/constants"
	myMiddleware "github.com/thalq/gopher_mart/internal/middleware"
	"github.com/thalq/gopher_mart/internal/orders"
	"github.com/thalq/gopher_mart/pkg/config"
	"github.com/thalq/gopher_mart/pkg/storage"
)

func NewRouter(cfg *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(myMiddleware.Logging)
	r.Use(myMiddleware.AuthMiddleware(constants.JWTSecret))

	db := storage.GetDB()
	authService := auth.NewAuthService(db, constants.JWTSecret)
	authHandler := auth.NewAuthHandler(authService)
	orderService := orders.NewOrderService(db)
	orderHandler := orders.NewOrderHandler(orderService, cfg.AccrualSystemAddress)
	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/orders", orderHandler.UploadOrder)
		r.Get("/orders", orderHandler.GetOrders)
		r.Get("/balance", orderHandler.GetBalance)
		r.Post("/balance/withdraw", orderHandler.WithdrawRequest)
		r.Get("/withdrawals", orderHandler.UserWithdrawls)
	})
	return r
}
