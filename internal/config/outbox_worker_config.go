package config

import "time"

type OutboxWorkerConfig struct {
	WorkerAttempts int8          `yaml:"workerAttempts" env:"WORKER_ATTEMPTS" env-default:"10"`
	WorkerInterval time.Duration `yaml:"workerInterval" env:"WORKER_INTERVAL" env-default:"1m" env-required:"true"`
}
