package worker

import (
	"context"
	"log"
	"time"

	"github.com/dmehra2102/budget-tracker/internal/config"
	"github.com/dmehra2102/budget-tracker/internal/repository"
	"github.com/dmehra2102/budget-tracker/internal/service"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CronWorker struct {
	cron         *cron.Cron
	cfg          *config.Config
	db           *mongo.Database
	alertService service.AlertService
	budgetRepo   repository.BudgetRepository
}

func NewCronWorker(
	cfg *config.Config,
	db *mongo.Database,
	alertService service.AlertService,
	budgetRepo repository.BudgetRepository,
) *CronWorker {
	return &CronWorker{
		cron:         cron.New(),
		cfg:          cfg,
		db:           db,
		alertService: alertService,
		budgetRepo:   budgetRepo,
	}
}

func (w *CronWorker) Start() error {
	_, err := w.cron.AddFunc(w.cfg.Worker.CleanupSchedule, w.cleanupJob)
	if err != nil {
		return err
	}

	_, err = w.cron.AddFunc(w.cfg.Worker.AlertCheckSchedule, w.alertCheckJob)
	if err != nil {
		return err
	}

	w.cron.Start()
	log.Println("Cron worker started")
	return nil
}

func (w *CronWorker) Stop() {
	w.cron.Stop()
	log.Println("Cron worker stopped")
}

func (w *CronWorker) cleanupJob() {
	ctx := context.Background()
	log.Println("Running cleanup job...")

	// Cleanup expired tokens
	cutoffDate := time.Now().AddDate(0, 0, -w.cfg.Worker.ExpiredTokenDays)
	_, err := w.db.Collection("users").UpdateMany(
		ctx,
		bson.M{"reset_token_expiry": bson.M{"$lt": cutoffDate}},
		bson.M{"$unset": bson.M{"reset_token": "", "reset_token_expiry": ""}},
	)

	if err != nil {
		log.Printf("Cleanup error: %v", err)
	}

	// Cleanup old alert notifications
	_, err = w.db.Collection("alert_notifications").DeleteMany(
		ctx,
		bson.M{"sent_at": bson.M{"$lt": cutoffDate}},
	)
	if err != nil {
		log.Printf("Cleanup error: %v", err)
	}

	log.Println("Cleanup job completed")
}

func (w *CronWorker) alertCheckJob() {
	ctx := context.Background()
	log.Println("Running alert check job...")

	if err := w.alertService.CheckAndSendAlerts(ctx); err != nil {
		log.Printf("Alert check error: %v", err)
	}

	log.Println("Alert check completed")
}
