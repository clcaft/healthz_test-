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
		App      `yaml:"app"`
		HTTP     `yaml:"http"`
		Log      `yaml:"logger"`
		Mongo    `yaml:"mongo"`
		MySQL    `yaml:"mysql"`
		RMQ      `yaml:"rabbitmq"`
		Services `yaml:"services"`
		Tracer   `yaml:"tracer"`
		GRPC
	}

	// App -.
	App struct {
		IsGoulash bool `env-required:"true" yaml:"is_goulash" env:"APP_IS_GOULASH" env-default:"true"`
	}

	// HTTP -.
	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	}
	GRPC struct {
		Port                   int  `env:"GRPC_PORT" env-default:"8081"`
		EnableServerReflection bool `env:"GRPC_SERVER_REFLECTION" env-default:"false"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" yaml:"log_level" env:"LOG_LEVEL"`
		Out   string `                    yaml:"log_out"   env:"LOG_OUT"`

		ElkRabbitUri string `yaml:"uri" env:"ELK_LOG_RABBITMQ"`
		ElkLevel     string `env:"ELK_LOG_LEVEL"`
		ElkExchange  string `env:"ELK_LOG_EXCHANGE" env-default:"log"`
		ElkCategory  string `yaml:"category" env:"ELASTIC_LOG_CATEGORY"`
		Partner      string `yaml:"partner" env:"PARTNER_NAME"`
	}

	Mongo struct {
		URI     string        `env-required:"true"                 env:"MONGO_DSN"`
		Timeout time.Duration `env-required:"true" yaml:"timeout"  env:"MONGO_TIMEOUT" env-default:"3s"`
	}

	MySQL struct {
		Username        string        `yaml:"username"          env:"DB_USERNAME"`
		Password        string        `yaml:"password"          env:"DB_PASSWORD"`
		Host            string        `yaml:"host"              env:"DB_HOST"`
		Port            string        `yaml:"port"              env:"DB_PORT"`
		DBName          string        `yaml:"db_name"           env:"DB_NAME"`
		Timeout         time.Duration `yaml:"timeout"           env:"DB_TIMEOUT"          env-default:"3s"`
		RefreshInterval time.Duration ` yaml:"refresh_interval"  env:"DB_REFRESH_INTERVAL" env-default:"5m"`
	}

	// RMQ -.
	RMQ struct {
		DsnList []string `env-required:"true" yaml:"rmq_dsn_list"      env:"RMQ_DSN_LIST"`
	}

	// Services Конфиг внешних сервисов.
	Services struct {
		MenuCatalog `yaml:"menu_catalog"`
	}

	// MenuCatalog Микросервис Каталога.
	MenuCatalog struct {
		ServiceURL string `yaml:"service_url" env:"MENU_CATALOG_SERVICE_URL"`
	}

	Tracer struct {
		TracerCollector string `env:"TRACER_COLLECTOR"`
		TracerService   string `env:"TRACER_SERVICE" env-default:"template-service"`
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
