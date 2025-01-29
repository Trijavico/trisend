package main

import (
	"trisend/internal/db"
	"trisend/internal/services"
)

type App struct {
	Auth         services.AuthService
	UserStore    db.UserStore
	SessionStore db.SessionStore
}
