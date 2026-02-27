package internal

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	PublicServerConfig PublicServerConfig
	PostgresConfig     PostgresConfig
	RedisConfig        RedisConfig
	AuthConfig         AuthConfig
}

type PublicServerConfig struct {
	Port string `env:"PUBLIC_SERVER_PORT" envDefault:"1323"`
}

type PostgresConfig struct {
	DBAdapter  string `env:"DB_ADAPTER"`
	DBName     string `env:"POSTGRES_DB"`
	DBHost     string `env:"POSTGRES_HOST"`
	DBPort     string `env:"POSTGRES_PORT"`
	DBUser     string `env:"POSTGRES_USER"`
	DBPassword string `env:"POSTGRES_PASSWORD"`
	DBSSLMode  string `env:"POSTGRES_SSLMODE"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST,required"`
	Port     int    `env:"REDIS_PORT" envDefault:"6379"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

type AuthConfig struct {
	JWTSecret          string `env:"JWT_SECRET" envDefault:"taskflow-dev-secret"`
	JWTExpirationHours int    `env:"JWT_EXPIRATION_HOURS" envDefault:"24"`
}

func NewConfig[T any](files ...string) (T, error) {
	// Загружаем .env файл, если он существует (игнорируем ошибку, если файла нет)
	_ = godotenv.Load(files...)

	cfg := *(new(T))
	return cfg, env.Parse(&cfg)
}
