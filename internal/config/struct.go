package config

import (
	"time"

	"github.com/guncv/ticket-reservation-server/internal/shared"
)

type Config struct {
	AppConfig      AppConfig      `mapstructure:"AppConfig"`
	DatabaseConfig DatabaseConfig `mapstructure:"DatabaseConfig"`
	RedisConfig    RedisConfig    `mapstructure:"RedisConfig"`
	AuthConfig     AuthConfig     `mapstructure:"AuthConfig"`
	TokenConfig    TokenConfig    `mapstructure:"TokenConfig"`
	BgJobsConfig   BgJobsConfig   `mapstructure:"BgJobsConfig"`
}

type AppConfig struct {
	AppPort     string   `mapstructure:"APP_PORT"`
	AppEnv      string   `mapstructure:"APP_ENV"`
	APIHost     string   `mapstructure:"API_HOST"`
	CORSOrigins []string `mapstructure:"CORS_ORIGINS"`
}

type DatabaseConfig struct {
	Host                  string        `mapstructure:"POSTGRES_HOST"`
	Port                  string        `mapstructure:"POSTGRES_PORT"`
	User                  string        `mapstructure:"POSTGRES_USER"`
	Password              string        `mapstructure:"POSTGRES_PASSWORD"`
	DbName                string        `mapstructure:"POSTGRES_DB"`
	ApplicationName       string        `mapstructure:"POSTGRES_APPLICATION_NAME"`
	SSLMode               string        `mapstructure:"POSTGRES_SSLMODE"`
	ConnectTimeout        time.Duration `mapstructure:"POSTGRES_CONNECT_TIMEOUT"`
	MaxOpenConns          int           `mapstructure:"POSTGRES_MAX_OPEN_CONNS"`
	MaxIdleConns          int           `mapstructure:"POSTGRES_MAX_IDLE_CONNS"`
	ConnMaxLifetime       time.Duration `mapstructure:"POSTGRES_CONN_MAX_LIFETIME"`
	ConnMaxLifetimeJitter time.Duration `mapstructure:"POSTGRES_CONN_MAX_LIFETIME_JITTER"`
	ConnMaxIdleTime       time.Duration `mapstructure:"POSTGRES_CONN_MAX_IDLE_TIME"`
	HealthCheckPeriod     time.Duration `mapstructure:"POSTGRES_HEALTH_CHECK_PERIOD"`
	EventTimeout          time.Duration `mapstructure:"POSTGRES_EVENT_TIMEOUT"`
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

type AuthConfig struct {
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	CookieDomain         string        `mapstructure:"COOKIE_DOMAIN"`
}

type TokenConfig struct {
	TokenType   string `mapstructure:"TOKEN_TYPE"`
	SecretKey   string `mapstructure:"TOKEN_SECRET_KEY"`
	TokenIssuer string `mapstructure:"TOKEN_ISSUER"`
}

type BgJobsConfig struct {
	TicketCounterInterval time.Duration `mapstructure:"TICKET_COUNTER_INTERVAL"`
}

const (
	// App defaults
	defaultAppEnv     = shared.AppEnvDev
	defaultAppPort    = "8080"
	defaultAPIHost    = "http://localhost:8080"
	defaultCORSOrigin = "http://localhost:3000"

	// Database defaults
	defaultPostgresHost                  = "localhost"
	defaultPostgresPort                  = "5432"
	defaultPostgresUser                  = "postgres"
	defaultPostgresPassword              = "postgres"
	defaultPostgresDB                    = "ticket-reservation-server"
	DefaultPostgresApplicationName       = "ticket-reservation-server"
	defaultPostgresSSLMode               = "disable"
	defaultPostgresConnectTimeout        = 10 * time.Second
	defaultPostgresMaxOpenConns          = 25
	defaultPostgresMaxIdleConns          = 5
	defaultPostgresConnMaxLifetime       = 5 * time.Minute
	defaultPostgresConnMaxLifetimeJitter = 30 * time.Second
	defaultPostgresConnMaxIdleTime       = 5 * time.Minute
	defaultPostgresHealthCheckPeriod     = 1 * time.Minute
	defaultPostgresEventTimeout          = 30 * time.Second

	// Redis defaults
	defaultRedisHost     = "localhost"
	defaultRedisPort     = "6379"
	defaultRedisPassword = ""
	defaultRedisDB       = 0

	// Token defaults
	defaultTokenType   = "jwt"
	defaultTokenIssuer = "ticket-reservation-server"

	// BgJobs defaults
	defaultTicketCounterInterval = 5 * time.Second
)

const (
	TokenTypeJWT    = "jwt"
	TokenTypePASETO = "paseto"
)
