package config

import (
	"os"
	"trisend/util"
)

var (
	APP_ENV     string
	SERVER_PORT string
	SSH_PORT    string
	DB_HOST     string
	DB_PORT     string
	DB_PASSWORD string
	DB_NAME     int

	JWT_SECRET     string
	SESSION_SECRET string

	SMTP_USER     string
	SMTP_PASSWORD string

	CLIENT_ID     string
	CLIENT_SECRET string
)

func LoadConfig() {
	APP_ENV = os.Getenv("APP_ENV")
	SERVER_PORT = util.GetEnvStr("PORT", "3000")
	SSH_PORT = util.GetEnvStr("SSH_PORT", "2222")

	DB_HOST = util.GetEnvStr("DB_HOST", "localhost")
	DB_PORT = util.GetEnvStr("DB_PORT", "6379")
	DB_PASSWORD = os.Getenv("DB_PASSWORD")
	DB_NAME = util.GetEnvInt("DB_NAME", 0)

	JWT_SECRET = os.Getenv("JWT_SECRET")
	SESSION_SECRET = os.Getenv("SESSION_SECRET")

	SMTP_USER = os.Getenv("SMTP_USERNAME")
	SMTP_PASSWORD = os.Getenv("SMTP_PASSWORD")

	CLIENT_ID = os.Getenv("CLIENT_ID")
	CLIENT_SECRET = os.Getenv("CLIENT_SECRET")
}

func IsAppEnvProd() bool {
	if APP_ENV == "dev" {
		return false
	}

	return true
}
