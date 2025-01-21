package server

import (
	"log/slog"
	"net/http"
	"trisend/db"
	"trisend/types"
	"trisend/util"
	"trisend/views"
	"trisend/views/components"
)

var create_sshkey_error = "Unable to register ssh key"

func handleKeysView(usrStore db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		value := r.Context().Value(SESSION_COOKIE)
		user := value.(*types.Session)

		userKeys, err := usrStore.GetSSHKeys(r.Context(), user.ID)
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

func handleCreateKey(usrStore db.UserStore) http.HandlerFunc {
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

		exists, err := usrStore.SSHKeyExists(r.Context(), fingerprint)
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

		err = usrStore.AddSSHKey(r.Context(), user.ID, title, fingerprint)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "Unable to add key", http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Redirect", "/keys")
		w.WriteHeader(http.StatusOK)
	}
}

func handleDeleteKey(usrStore db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sshID := r.PathValue("id")

		err := usrStore.DeleteSSHKey(r.Context(), sshID)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "Unable to delete key", http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
	}
}
