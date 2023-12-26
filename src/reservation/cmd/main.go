package main

import (
	"context"
	"fmt"
	store "go.example/saga/pkg/store/postgres"
	"go.example/saga/reservation/internal/controller/reservation"
	httphandler "go.example/saga/reservation/internal/handler/http"
	"go.example/saga/reservation/internal/handler/ingester/kafka"
	"go.example/saga/reservation/internal/repository/postgres"
	"go.example/saga/reservation/pkg/model"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
)

const serviceName = "reservation"

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

	//cfg := InMem()

	logger.Info(fmt.Sprintf("Starting the %s service", serviceName), zap.Int("port", cfg.Server.Port))

	addr := cfg.Kafka.BoostrapServers
	rbGroupID := cfg.Kafka.RoomBooking.GroupID
	rbTopic := cfg.Kafka.RoomBooking.InboxTopic
	roomBookIngester, err := kafka.NewIngester[model.BookingEventPayload](addr, rbGroupID, rbTopic)
	if err != nil {
		logger.Fatal("Failed to init room booking kafka ingester", zap.Error(err))
		return
	}

	pGroupID := cfg.Kafka.Payment.GroupID
	pTopic := cfg.Kafka.Payment.InboxTopic
	paymentIngester, err := kafka.NewIngester[model.PaymentEventPayload](addr, pGroupID, pTopic)
	if err != nil {
		logger.Fatal("Failed to init payment kafka ingester", zap.Error(err))
		return
	}

	st, err := store.NewStore(cfg.Store.StoreProps())
	if err != nil {
		logger.Fatal("Failed to open postgres configs", zap.Error(err))
	}

	eventLogger := store.NewEventLogs()
	sagaSagaRepository := store.NewSagaRepository()
	repository := postgres.New()
	ctrl := reservation.New(st, eventLogger, repository, sagaSagaRepository, roomBookIngester, paymentIngester)

	ctx := context.Background()
	go func() {
		err = ctrl.StartBookingIngestion(ctx)
		if err != nil {
			logger.Fatal("Failed to start kafka room booking ingester", zap.Error(err))
		}
	}()

	go func() {
		err = ctrl.StartPaymentIngestion(ctx)
		if err != nil {
			logger.Fatal("Failed to start kafka payment ingester", zap.Error(err))
		}
	}()

	h := httphandler.New(ctrl)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port), h); err != nil {
		panic(err)
	}
}
