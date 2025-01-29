package main

import (
	"html/template"
	"trisend/internal/db"
	"trisend/internal/services"
)

type App struct {
	Auth             services.AuthService
	UserStore        db.UserStore
	SessionStore     db.SessionStore
	AuthCodeTemplate *template.Template
}
