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
		Server: serverConfig{Port: 8082},
		Store: storeConfig{
			Host:     "localhost",
			Port:     "5434",
			User:     "paymentuser",
			Password: "secret",
			Dbname:   "paymentdb",
		},
		Kafka: kafkaConfig{
			BoostrapServers: "localhost:29092",
			Payment: sagaConfig{
				GroupID:    "payment-service",
				InboxTopic: "payment.inbox.events",
			},
		},
	}
}
