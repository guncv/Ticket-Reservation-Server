package config

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppConfig      AppConfig      `mapstructure:"AppConfig"`
	DatabaseConfig DatabaseConfig `mapstructure:"DatabaseConfig"`
	RedisConfig    RedisConfig    `mapstructure:"RedisConfig"`
}

type AppConfig struct {
	AppPort     string   `mapstructure:"APP_PORT"`
	AppEnv      string   `mapstructure:"APP_ENV"`
	APIHost     string   `mapstructure:"API_HOST"`
	CORSOrigins []string `mapstructure:"CORS_ORIGINS"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"POSTGRES_HOST"`
	Port     string `mapstructure:"POSTGRES_PORT"`
	User     string `mapstructure:"POSTGRES_USER"`
	Password string `mapstructure:"POSTGRES_PASSWORD"`
	DbName   string `mapstructure:"POSTGRES_DB"`
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DBTemp   int    `mapstructure:"REDIS_DB_TEMP"`
	DBQueue  int    `mapstructure:"REDIS_DB_QUEUE"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if path := os.Getenv("CONFIG_PATH"); path != "" {
		v.SetConfigFile(path)
	} else {
		env := os.Getenv("ENV")
		if env == "" {
			env = "dev"
		}
		v.SetConfigName("config." + env)
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
	}

	if err := v.ReadInConfig(); err != nil {
		log.Printf("[config] No config file found, using ENV/defaults only: %v", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
