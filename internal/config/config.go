package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"time"
)

// Config представляет конфигурацию приложения
type Config struct {
	Env                 string        `yaml:"ENV" env:"ENV" env-default:"production"`
	Address             string        `yaml:"address" env:"ADDRESS" env-default:"localhost:8080"`
	DbHost              string        `yaml:"db_host" env:"DB_HOST" env-required:"true" `
	DbPort              string        `yaml:"db_port" env:"DB_PORT" env-required:"true"`
	DbName              string        `yaml:"db_name" env:"DB_NAME" env-required:"true"`
	DbUser              string        `yaml:"db_user" env:"DB_USER" env-required:"true"`
	DbPassword          string        `yaml:"db_password" env:"DB_PASSWORD" env-required:"true"`
	DbMaxConnections    int32         `yaml:"db_max_connections" env:"DB_MAX_CONNECTIONS"`
	DbMinConnections    int32         `yaml:"db_min_connections" env:"DB_MIN_CONNECTIONS"`
	DbMaxConnLifetime   time.Duration `yaml:"db_max_conn_lifetime" env:"DB_MAX_CONN_LIFETIME"`
	DbMaxConnIdleTime   time.Duration `yaml:"db_max_conn_idle_time" env:"DB_MAX_CONN_IDLE_TIME"`
	DbHealthCheckPeriod time.Duration `yaml:"db_health_check_period" env:"DB_HEALTH_CHECK_PERIOD"`
	JWTSecretKey        string        `yaml:"jwt_secret_key" env:"JWT_SECRET_KEY" env-required:"true"`
	JWTDuration         time.Duration `yaml:"jwt_duration"  env:"JWT_DURATION" env-default:"5m"`
	ServerTimeout       time.Duration `yaml:"server_timeout" env:"SERVER_TIMEOUT" env-default:"10s"`
}

// LoadConfig загружает конфигурацию из файла и переменных окружения
func LoadConfig(filePath string) (*Config, error) {
	var cfg Config
	// Читаем основной конфиг файл
	err := cleanenv.ReadConfig("config.yaml", &cfg)
	if err != nil {
		log.Default().Printf("Error loading config from base config file: %s", err.Error())
	}
	// читаем конфиг файл с файла с секретами
	err = cleanenv.ReadConfig(filePath, &cfg)
	if err != nil {
		log.Default().Printf("Error loading config from secret config file: %s", err.Error())
	}
	//В конце смотрим в переменные окружения
	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Default().Printf("Error loading config from env vars: %s", err.Error())
		return nil, err
	}
	return &cfg, nil
}
