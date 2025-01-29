package main

import (
	"log/slog"
	"net/http"
	"trisend/internal/types"
	"trisend/internal/util"
	"trisend/internal/views"
	"trisend/internal/views/components"
)

var create_sshkey_error = "Unable to register ssh key"

func handleKeysView(app App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		value := r.Context().Value(SESSION_COOKIE)
		user := value.(*types.Session)

		userKeys, err := app.UserStore.GetSSHKeys(r.Context(), user.ID)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "Faile to get keys", http.StatusInternalServerError)
			return
		}

		profile := components.ProfileButton(user)
		views.Keys(profile, userKeys).Render(r.Context(), w)
	}
}

func handleCreateKeyView() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(SESSION_COOKIE).(*types.Session)
		profile := components.ProfileButton(user)

		validation := types.ValidationSSHForm{
			Fields: map[string]string{},
			Errors: map[string]string{},
		}
		views.SSHKeyForm(profile, validation).Render(r.Context(), w)
	}
}

func handleCreateKey(app App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		validation := types.ValidationSSHForm{
			Fields: map[string]string{},
			Errors: map[string]string{},
		}
		user := r.Context().Value(SESSION_COOKIE).(*types.Session)
		title := r.FormValue("title")
		key := r.FormValue("key")

		ok := validation.Validate(title, key)
		if !ok {
			views.CreateSSHForm(validation).Render(r.Context(), w)
			return
		}

		fingerprint, err := util.GetFingerPrint(key)
		if err != nil {
			validation.Errors["key"] = "Invalid key"
			views.CreateSSHForm(validation).Render(r.Context(), w)
			return
		}

		exists, err := app.UserStore.SSHKeyExists(r.Context(), fingerprint)
		if err != nil {
			validation.Errors["key"] = "Key already exists"
			views.CreateSSHForm(validation).Render(r.Context(), w)
			return
		}

		if exists {
			validation.Errors["key"] = "SSH Key already exists"
			views.CreateSSHForm(validation).Render(r.Context(), w)
			return
		}

		err = app.UserStore.AddSSHKey(r.Context(), user.ID, title, fingerprint)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "Unable to add key", http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Redirect", "/keys")
		w.WriteHeader(http.StatusOK)
	}
}

func handleDeleteKey(app App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sshID := r.PathValue("id")

		err := app.UserStore.DeleteSSHKey(r.Context(), sshID)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "Unable to delete key", http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
	}
}
