package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"time"
)

// Config представляет конфигурацию приложения
type Config struct {
	Env          string        `yaml:"ENV" env:"ENV" env-default:"production"`
	Address      string        `yaml:"address" env:"ADDRESS" env-default:"localhost:8080"`
	DbHost       string        `yaml:"db_host" env:"DB_HOST" env-required:"true" `
	DbPort       string        `yaml:"db_port" env:"DB_PORT" env-required:"true"`
	DbName       string        `yaml:"db_name" env:"DB_NAME" env-required:"true"`
	DbUser       string        `yaml:"db_user" env:"DB_USER" env-required:"true"`
	DbPassword   string        `yaml:"db_password" env:"DB_PASSWORD" env-required:"true"`
	JWTSecretKey string        `yaml:"jwt_secret_key" env:"JWT_SECRET_KEY" env-required:"true"`
	JWTDuration  time.Duration `yaml:"jwt_duration"  env:"JWT_DURATION" env-default:"5m"`
}

// LoadConfig загружает конфигурацию из файла и переменных окружения
func LoadConfig(filePath string) (*Config, error) {
	var cfg Config
	//Сначала смотрим в переменные окружения
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Default().Printf("Error loading config from env vars: %s", err.Error())
	} else {
		return &cfg, nil
	}

	err = cleanenv.ReadConfig(filePath, &cfg)
	if err != nil {
		log.Default().Printf("Error loading config from file: %s", err.Error())
	} else {
		return &cfg, nil
	}
	return nil, err
}
