package server

import (
	"net/http"
	"trisend/db"
	"trisend/types"
	"trisend/util"
	"trisend/views"
	"trisend/views/components"
)

func handleKeysView(usrStore db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		value := r.Context().Value(SESSION_COOKIE)
		user := value.(*types.Session)

		userKeys, err := usrStore.GetSSHKeys(r.Context(), user.ID)
		if err != nil {
			handleError(w, "Failed to get keys", err, http.StatusInternalServerError)
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
		views.SSHKeyForm(profile).Render(r.Context(), w)
	}
}

func handleCreateKey(usrStore db.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(SESSION_COOKIE).(*types.Session)
		title := r.FormValue("title")
		key := r.FormValue("key")

		fingerprint, err := util.GetFingerPrint(key)
		if err != nil {
			handleError(w, "an error occurred", err, http.StatusInternalServerError)
			return
		}

		err = usrStore.AddSSHKey(r.Context(), user.ID, title, fingerprint)
		if err != nil {
			http.Error(w, "an error occurred", http.StatusInternalServerError)
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
			http.Error(w, "an error occurred", http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
	}
}
