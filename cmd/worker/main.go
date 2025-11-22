package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dmehra2102/budget-tracker/internal/config"
	"github.com/dmehra2102/budget-tracker/internal/database"
	"github.com/dmehra2102/budget-tracker/internal/repository"
	"github.com/dmehra2102/budget-tracker/internal/service"
	"github.com/dmehra2102/budget-tracker/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewMongoDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB())
	budgetRepo := repository.NewBudgetRepository(db.DB())
	alertRepo := repository.NewAlertRepository(db.DB())

	// Initialize Services
	emailService := service.NewEmailService(cfg)
	alertService := service.NewAlertService(alertRepo, budgetRepo, userRepo, emailService)

	cronWorker := worker.NewCronWorker(cfg, db.DB(), alertService, budgetRepo)

	if err := cronWorker.Start(); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	log.Println("Worker Started Successfully")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down worker...")
	cronWorker.Stop()
	log.Println("Worker exited")
}
