package config

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/guncv/ticket-reservation-server/internal/shared"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func LoadConfig(flags *pflag.FlagSet) (*Config, error) {
	_ = godotenv.Load("../.env")
	v := viper.New()

	// Set defaults
	setDefaults(v)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind env to app environment
	v.BindEnv("AppConfig.APP_ENV", "APP_ENV")

	env := v.GetString("AppConfig.APP_ENV")

	v.SetConfigName(env)
	v.SetConfigType("yaml")
	v.AddConfigPath("../../config")
	v.AddConfigPath("./config")
	v.AddConfigPath("../config")

	if err := v.ReadInConfig(); err != nil {
		if env != shared.AppEnvTest {
			log.Printf("[config] No config file found, using ENV/defaults only: %v", err)
		}
	}

	// Bind env to config
	v.BindEnv("AuthConfig.ACCESS_TOKEN_DURATION", "ACCESS_TOKEN_DURATION")
	v.BindEnv("AuthConfig.REFRESH_TOKEN_DURATION", "REFRESH_TOKEN_DURATION")
	v.BindEnv("AuthConfig.COOKIE_DOMAIN", "COOKIE_DOMAIN")
	v.BindEnv("TokenConfig.TOKEN_TYPE", "TOKEN_TYPE")
	v.BindEnv("TokenConfig.TOKEN_SECRET_KEY", "TOKEN_SECRET_KEY")
	v.BindEnv("TokenConfig.TOKEN_ISSUER", "TOKEN_ISSUER")

	// Bind flags to config
	if flags != nil {
		bindFlagsToConfig(flags, v)
		if err := v.BindPFlags(flags); err != nil {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func bindFlagsToConfig(flags *pflag.FlagSet, v *viper.Viper) {
	flagMap := map[string]string{
		"app-port": "AppConfig.APP_PORT",
	}

	flags.VisitAll(func(flag *pflag.Flag) {
		if configPath, exists := flagMap[flag.Name]; exists {
			_ = v.BindPFlag(configPath, flag)
		}
	})
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("AppConfig.APP_PORT", defaultAppPort)
	v.SetDefault("AppConfig.APP_ENV", defaultAppEnv)
	v.SetDefault("AppConfig.API_HOST", defaultAPIHost)
	v.SetDefault("AppConfig.CORS_ORIGINS", []string{defaultCORSOrigin})

	v.SetDefault("DatabaseConfig.POSTGRES_HOST", defaultPostgresHost)
	v.SetDefault("DatabaseConfig.POSTGRES_PORT", defaultPostgresPort)
	v.SetDefault("DatabaseConfig.POSTGRES_USER", defaultPostgresUser)
	v.SetDefault("DatabaseConfig.POSTGRES_PASSWORD", defaultPostgresPassword)
	v.SetDefault("DatabaseConfig.POSTGRES_DB", defaultPostgresDB)
	v.SetDefault("DatabaseConfig.POSTGRES_APPLICATION_NAME", DefaultPostgresApplicationName)
	v.SetDefault("DatabaseConfig.POSTGRES_SSLMODE", defaultPostgresSSLMode)
	v.SetDefault("DatabaseConfig.POSTGRES_CONNECT_TIMEOUT", defaultPostgresConnectTimeout)
	v.SetDefault("DatabaseConfig.POSTGRES_MAX_OPEN_CONNS", defaultPostgresMaxOpenConns)
	v.SetDefault("DatabaseConfig.POSTGRES_MAX_IDLE_CONNS", defaultPostgresMaxIdleConns)
	v.SetDefault("DatabaseConfig.POSTGRES_CONN_MAX_LIFETIME", defaultPostgresConnMaxLifetime)
	v.SetDefault("DatabaseConfig.POSTGRES_CONN_MAX_LIFETIME_JITTER", defaultPostgresConnMaxLifetimeJitter)
	v.SetDefault("DatabaseConfig.POSTGRES_CONN_MAX_IDLE_TIME", defaultPostgresConnMaxIdleTime)
	v.SetDefault("DatabaseConfig.POSTGRES_HEALTH_CHECK_PERIOD", defaultPostgresHealthCheckPeriod)
	v.SetDefault("DatabaseConfig.POSTGRES_EVENT_TIMEOUT", defaultPostgresEventTimeout)

	v.SetDefault("RedisConfig.REDIS_HOST", defaultRedisHost)
	v.SetDefault("RedisConfig.REDIS_PORT", defaultRedisPort)
	v.SetDefault("RedisConfig.REDIS_PASSWORD", defaultRedisPassword)
	v.SetDefault("RedisConfig.REDIS_DB", defaultRedisDB)

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = defaultAppEnv
	}
	if appEnv == shared.AppEnvTest {
		v.SetDefault("TokenConfig.TOKEN_SECRET_KEY", shared.TestTokenSecretKey)
	}

	v.SetDefault("TokenConfig.TOKEN_TYPE", defaultTokenType)
	v.SetDefault("TokenConfig.TOKEN_ISSUER", defaultTokenIssuer)
}

func (c *Config) Validate() error {

	if c.AppConfig.AppPort == "" {
		return errors.New("APP_PORT is required")
	}
	if c.AppConfig.AppEnv == "" {
		return errors.New("APP_ENV is required")
	}

	if c.AppConfig.AppEnv != shared.AppEnvDev && c.AppConfig.AppEnv != shared.AppEnvProd && c.AppConfig.AppEnv != shared.AppEnvTest {
		return errors.New("APP_ENV must be " + shared.AppEnvDev + ", " + shared.AppEnvProd + ", or " + shared.AppEnvTest)
	}

	if c.AppConfig.APIHost == "" {
		return errors.New("API_HOST is required")
	}

	if len(c.AppConfig.CORSOrigins) == 0 {
		return errors.New("CORS_ORIGINS is required")
	}

	if c.DatabaseConfig.Host == "" {
		return errors.New("POSTGRES_HOST is required")
	}

	if c.DatabaseConfig.Port == "" {
		return errors.New("POSTGRES_PORT is required")
	}

	if c.DatabaseConfig.User == "" {
		return errors.New("POSTGRES_USER is required")
	}

	if c.DatabaseConfig.DbName == "" {
		return errors.New("POSTGRES_DB is required")
	}

	if c.TokenConfig.SecretKey == "" {
		return errors.New("TOKEN_SECRET_KEY is required")
	}
	if c.TokenConfig.TokenType == "" {
		return errors.New("TOKEN_TYPE is required")
	}
	if c.TokenConfig.TokenType != TokenTypeJWT && c.TokenConfig.TokenType != TokenTypePASETO {
		return errors.New("TOKEN_TYPE must be jwt or paseto")
	}
	if c.TokenConfig.TokenIssuer == "" {
		return errors.New("TOKEN_ISSUER is required")
	}

	return nil
}
