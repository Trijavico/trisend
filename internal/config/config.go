package config

import (
	"fmt"
	"os"
	"strconv"
)

var (
	APP_ENV     string
	SERVER_PORT string
	SSH_PORT    string
	HOST        string
	DOMAIN_NAME string

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
	SERVER_PORT = os.Getenv("PORT")
	SSH_PORT = os.Getenv("SSH_PORT")
	DOMAIN_NAME = os.Getenv("HOST")
	if APP_ENV == "dev" {
		HOST = fmt.Sprintf("http://%s", DOMAIN_NAME)
	} else {
		HOST = fmt.Sprintf("https://%s", DOMAIN_NAME)
	}

	DB_HOST = os.Getenv("DB_HOST")
	DB_PORT = os.Getenv("DB_PORT")
	DB_PASSWORD = os.Getenv("DB_PASSWORD")

	db_name := os.Getenv("DB_NAME")
	DB_NAME, _ = strconv.Atoi(db_name)

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
