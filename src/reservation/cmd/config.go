package main

import "go.example/saga/pkg/store/postgres"

type (
	config struct {
		Server serverConfig `yaml:"server"`
		Store  storeConfig  `yaml:"store"`
		Kafka  kafkaConfig  `yaml:"kafka"`
	}

	serverConfig struct {
		Port int `yaml:"port"`
	}

	storeConfig struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Dbname   string `yaml:"dbname"`
	}

	kafkaConfig struct {
		BoostrapServers string     `yaml:"boostrap-servers"`
		RoomBooking     sagaConfig `yaml:"room-booking"`
		Payment         sagaConfig `yaml:"payment"`
	}

	sagaConfig struct {
		GroupID    string `yaml:"group-id"`
		InboxTopic string `yaml:"inbox-topic"`
	}
)

func (s storeConfig) StoreProps() postgres.StoreProps {
	return postgres.StoreProps{
		Host:     s.Host,
		Port:     s.Port,
		User:     s.User,
		Password: s.Password,
		Dbname:   s.Dbname,
	}
}

func InMem() config {
	return config{
		Server: serverConfig{Port: 8080},
		Store: storeConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "reservationuser",
			Password: "secret",
			Dbname:   "reservationdb",
		},
		Kafka: kafkaConfig{
			BoostrapServers: "localhost:29092",
			RoomBooking: sagaConfig{
				GroupID:    "reservation-service-rb",
				InboxTopic: "room-booking.outbox.events",
			},
			Payment: sagaConfig{
				GroupID:    "reservation-service-p",
				InboxTopic: "payment.outbox.events",
			},
		},
	}
}
