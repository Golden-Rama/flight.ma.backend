package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppName       string `envconfig:"APP_NAME" default:"GR Multi Aggregator"`
	AppEnv        string `envconfig:"APP_ENV" default:"local"`
	AppDebug      bool   `envconfig:"APP_DEBUG" default:"true"`
	ApiPort       int    `envconfig:"API_PORT" default:"4001"`
	DashboardPort int    `envconfig:"DASHBOARD_PORT" default:"4000"`

	DBDriver string `envconfig:"DB_DRIVER" default:"mysql"`
	DBHost   string `envconfig:"DB_HOST" default:"127.0.0.1"`
	DBPort   int    `envconfig:"DB_PORT" default:"3306"`
	DBUser   string `envconfig:"DB_USER" default:"root"`
	DBPass   string `envconfig:"DB_PASS" default:""`
	DBName   string `envconfig:"DB_NAME" default:"flight_db"`

	JwtSecret string `envconfig:"JWT_SECRET" default:"super_secret_jwt_key"`
}

var data *Config

func Get() *Config {
	if data == nil {
		_ = godotenv.Load() // optional loading of .env
		data = &Config{}
		envconfig.MustProcess("", data)
	}
	return data
}
