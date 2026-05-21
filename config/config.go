package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type (
	// Config -.
	Config struct {
		App    `yaml:"app"`
		HTTP   `yaml:"http"`
		GRPC   `yaml:"grpc"`
		Log    `yaml:"logger"`
		MySQL  `yaml:"mysql"`
		RMQ    `yaml:"rabbitmq"`
		Tracer `yaml:"tracer"`
	}

	// App -.
	App struct {
		Name    string `yaml:"name"    env:"APP_NAME"    env-default:"go-service"`
		Version string `yaml:"version" env:"APP_VERSION" env-default:"0.0.1"`
	}

	// HTTP -.
	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	}

	GRPC struct {
		Port                   int  `yaml:"port"                     env:"GRPC_PORT"              env-default:"8081"`
		EnableServerReflection bool `yaml:"enable_server_reflection" env:"GRPC_SERVER_REFLECTION" env-default:"false"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" yaml:"log_level" env:"LOG_LEVEL"`
		Out   string `yaml:"log_out"   env:"LOG_OUT" env-default:"console"`

		ElkRabbitUri string `yaml:"uri" env:"ELK_LOG_RABBITMQ"`
		ElkLevel     string `env:"ELK_LOG_LEVEL"`
		ElkExchange  string `env:"ELK_LOG_EXCHANGE" env-default:"log"`
		ElkCategory  string `yaml:"category" env:"ELASTIC_LOG_CATEGORY"`
		Partner      string `yaml:"partner" env:"PARTNER_NAME"`
	}

	MySQL struct {
		Username        string        `yaml:"username"         env:"DB_USERNAME"         env-default:"root"`
		Password        string        `yaml:"password"         env:"DB_PASSWORD"`
		Host            string        `yaml:"host"             env:"DB_HOST"             env-default:"localhost"`
		Port            string        `yaml:"port"             env:"DB_PORT"             env-default:"3306"`
		DBName          string        `yaml:"db_name"          env:"DB_NAME"`
		Charset         string        `yaml:"charset"          env:"DB_CHARSET"          env-default:"utf8mb4"`
		SlaveCount      int           `yaml:"slave_count"      env:"DB_SLAVE_COUNT"      env-default:"1"`
		Timeout         time.Duration `yaml:"timeout"          env:"DB_TIMEOUT"          env-default:"3s"`
		RefreshInterval time.Duration `yaml:"refresh_interval" env:"DB_REFRESH_INTERVAL" env-default:"5m"`
	}

	// RMQ -.
	RMQ struct {
		DsnList      []string          `yaml:"dsn_list"      env:"RMQ_DSN_LIST"`
		ExchangeName string            `yaml:"exchange_name" env:"RMQ_EXCHANGE_NAME" env-default:"go-service"`
		ExchangeType string            `yaml:"exchange_type" env:"RMQ_EXCHANGE_TYPE" env-default:"direct"`
		Queues       map[string]string `yaml:"queues"`
	}

	Tracer struct {
		TracerCollector string `yaml:"collector" env:"TRACER_COLLECTOR"`
		TracerService   string `yaml:"service"   env:"TRACER_SERVICE" env-default:"go-service"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	// Загрузка переменных окружения
	_ = godotenv.Load(".env")

	err := cleanenv.ReadConfig("./config/config.yml", cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}