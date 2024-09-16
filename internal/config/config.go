package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env           string `env:"ENV" env-default:"local"`
	Postgres      PostgresConfig
	Redis         RedisConfig
	Server        ServerConfig
	TemplatesPath string `env:"TEMPLATES_PATH" env-required:"true"`
	Auth          AuthConfig
}

type ServerConfig struct {
	Port string `yaml:"port" env:"PORT"  env-default:"8080"`
	//Host            string        `yaml:"host" env:"HOST" env-default:"localhost"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" env-default:"30s"`
	WriteTimeout    time.Duration `yaml:"wtite_timeout" env:"WRITE_TIMEOUT" env-default:"30s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" env-default:"10s"`
}

type PostgresConfig struct {
	PostgresURL string `env:"POSTGRES_URL" env-required:"true"`
}

type RedisConfig struct {
	Hosts    []string `env:"REDIS_HOSTS" yaml:"hosts" env-required:"true"`
	Password string   `env:"REDIS_PASSWORD" env-required:"true"`
}

type AuthConfig struct {
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" env-default:"15m"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env-default:"720h"`
	PasswordSalt    string        `env:"PASSWORD_SALT" env-required:"true"`
	JWTSigningKey   string        `env:"JWT_SIGNING_KEY" env-required:"true"`
}

func InitConfig() (*Config, error) {
	path := fetchConfigPath()

	if path == "" {
		return nil, fmt.Errorf("'.env' file path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exists: " + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("can not read config and parse it: %w", err)
	}

	return &cfg, nil
}
func fetchConfigPath() string {
	var path string

	flag.StringVar(&path, "env", "", "application configuration file")
	flag.Parse()

	if path == "" {
		path = os.Getenv("ENV_PATH")
	}

	return path
}
