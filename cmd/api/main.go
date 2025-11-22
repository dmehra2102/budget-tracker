package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmehra2102/budget-tracker/internal/cache"
	"github.com/dmehra2102/budget-tracker/internal/config"
	"github.com/dmehra2102/budget-tracker/internal/database"
	"github.com/dmehra2102/budget-tracker/internal/handler"
	"github.com/dmehra2102/budget-tracker/internal/middleware"
	"github.com/dmehra2102/budget-tracker/internal/repository"
	"github.com/dmehra2102/budget-tracker/internal/service"
	"github.com/dmehra2102/budget-tracker/internal/utils"
	"github.com/gorilla/mux"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize MongoDB
	db, err := database.NewMongoDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer db.Close()

	redisClient, err := cache.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	jwtAuth := utils.NewJWTAuth(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL)

	cacheService := cache.NewCacheService(redisClient, cfg)

	userRepo := repository.NewUserRepository(db.DB())
	// alertRepo := repository.NewAlertRepository(db.DB())
	budgetRepo := repository.NewBudgetRepository(db.DB())

	emailService := service.NewEmailService(cfg)
	authService := service.NewAuthService(userRepo, emailService, cfg, jwtAuth)
	budgetService := service.NewBudgetService(budgetRepo, cacheService)
	// alertService := service.NewAlertService(alertRepo, budgetRepo, userRepo, emailService)

	authHandler := handler.NewAuthHandler(authService)
	budgetHandler := handler.NewBudgetHandler(budgetService)

	router := setupRouter(jwtAuth, authHandler, budgetHandler)

	// Create server
	srv := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRouter(
	jwtAuth *utils.JWTAuth,
	authHandler *handler.AuthHandler,
	budgetHandler *handler.BudgetHandler,

) *mux.Router {
	router := mux.NewRouter()

	// Global middlewares
	router.Use(middleware.CORSMiddleware)
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.ErrorHandlerMiddleware)
	router.Use(middleware.TimeoutMiddleware(30 * time.Second))

	rateLimiter := middleware.NewRateLimiter(100, 20)
	router.Use(rateLimiter.Middleware)

	// Public routes
	router.HandleFunc("/api/v1/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/v1/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/v1/auth/forgot-password", authHandler.ForgotPassword).Methods("POST")
	router.HandleFunc("/api/v1/auth/reset-password", authHandler.ResetPassword).Methods("POST")

	protected := router.PathPrefix("/api/v1").Subrouter()
	protected.Use(middleware.AuthMiddleware(jwtAuth))

	// Budget routes
	protected.HandleFunc("/budgets", budgetHandler.CreateBudget).Methods("POST")
	protected.HandleFunc("/budgets", budgetHandler.GetBudgets).Methods("GET")
	protected.HandleFunc("/budgets/{id}", budgetHandler.GetBudget).Methods("GET")
	protected.HandleFunc("/budgets/{id}", budgetHandler.UpdateBudget).Methods("PUT")
	protected.HandleFunc("/budgets/{id}", budgetHandler.DeleteBudget).Methods("DELETE")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	return router
}
