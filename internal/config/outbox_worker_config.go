package config

import "time"

type OutboxWorkerConfig struct {
	WorkerAttempts int           `yaml:"workerAttempts" env:"WORKER_ATTEMPTS" env-default:"10" env-required:"true"`
	WorkerInterval time.Duration `yaml:"workerInterval" env:"WORKER_INTERVAL" env-default:"1m" env-required:"true"`
}
