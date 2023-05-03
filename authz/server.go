package authz

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/auth"
)

func NewServer(checker *Checker) (http.Handler, error) {
	r := chi.NewRouter()
	r.Get("/users", wrapHandler(usersHandler, checker))
	r.Get("/users/{id}", wrapHandler(userHandler, checker))
	r.Get("/feeds/", wrapHandler(listFeedsHandler, checker))
	r.Get("/feeds/{id}/permissions", wrapHandler(feedPermissionsHandler, checker))
	r.Get("/feed_versions/", wrapHandler(listFeedVersionsHandler, checker))
	r.Get("/feed_versions/{id}/permissions", wrapHandler(feedVersionPermissionsHandler, checker))
	return r, nil
}

func wrapHandler(next func(http.ResponseWriter, *http.Request, *Checker), checker *Checker) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.ForContext(r.Context())
		if user == nil {
			http.Error(w, "not logged in", http.StatusUnauthorized)
			return
		}
		next(w, r, checker)
	})
}

func usersHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	users, err := checker.authn.Users(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		log.Error().Err(err).Msg("authn.Users")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	jj, _ := json.Marshal(users)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jj)
}

func userHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	user, err := checker.authn.UserByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		log.Error().Err(err).Msg("authn.UserByID")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	jj, _ := json.Marshal(user)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jj)
}

func listFeedsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.ListFeeds(
		r.Context(),
		auth.ForContext(r.Context()),
	)
	if err != nil {
		log.Error().Err(err).Msg("checker.ListFeeds")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jj, _ := json.Marshal(ret)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jj)
}

func listFeedVersionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.ListFeeds(
		r.Context(),
		auth.ForContext(r.Context()),
	)
	if err != nil {
		log.Error().Err(err).Msg("checker.ListFeeds")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jj, _ := json.Marshal(ret)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jj)
}

func feedPermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.FeedPermissions(
		r.Context(),
		auth.ForContext(r.Context()),
		atoi(chi.URLParam(r, "id")),
	)
	if err != nil {
		log.Error().Err(err).Msg("checker.FeedPermissions")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("ret:", ret)
	jj, _ := json.Marshal(ret)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jj)
}

func feedVersionPermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.FeedVersionPermissions(r.Context(), auth.ForContext(r.Context()), atoi(chi.URLParam(r, "id")))
	if err != nil {
		log.Error().Err(err).Msg("checker.FeedVersionPermissions")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jj, _ := json.Marshal(ret)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jj)
}
