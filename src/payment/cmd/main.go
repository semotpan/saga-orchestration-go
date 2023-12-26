package main

import (
	"context"
	"fmt"
	"go.example/saga/payment/internal/controller/payment"
	"go.example/saga/payment/internal/ingester/kafka"
	"go.example/saga/payment/internal/repository/postgres"
	store "go.example/saga/pkg/store/postgres"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
)

const serviceName = "hotel"

func main() {
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	f, err := os.Open("app.yaml")
	if err != nil {
		logger.Fatal("Failed to open configuration", zap.Error(err))
	}

	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to parse configuration", zap.Error(err))
	}

	logger.Info(fmt.Sprintf("Starting the %s service", serviceName))

	roomBookIngester, err := kafka.NewIngester(
		cfg.Kafka.BoostrapServers, cfg.Kafka.Payment.GroupID, cfg.Kafka.Payment.InboxTopic)
	if err != nil {
		logger.Fatal("Failed to init room booking kafka ingester", zap.Error(err))
		return
	}

	st, err := store.NewStore(cfg.Store.StoreProps())
	if err != nil {
		logger.Fatal("Failed to open postgres configs", zap.Error(err))
	}

	repository := postgres.New()
	eventLogger := store.NewEventLogs()
	ctrl := payment.New(st, repository, roomBookIngester, eventLogger)

	ctx := context.Background()
	err = ctrl.StartIngestion(ctx)
	if err != nil {
		logger.Fatal("Failed to start kafka ingester", zap.Error(err))
	}
}
