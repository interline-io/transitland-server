package authz

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/auth"
)

func NewServer(checker *Checker) (http.Handler, error) {
	r := chi.NewRouter()
	r.Get("/users", wrapHandler(userIndexHandler, checker))
	r.Get("/users/{id}", wrapHandler(userPermissionsHandler, checker))

	r.Get("/tenants", wrapHandler(tenantIndexHandler, checker))
	r.Get("/tenants/{id}", wrapHandler(tenantPermissionsHandler, checker))

	r.Get("/groups", wrapHandler(groupIndexHandler, checker))
	r.Get("/groups/{id}", wrapHandler(groupPermissionsHandler, checker))
	r.Post("/groups/{id}/permissions/{relation}/{user}", wrapHandler(groupAddPermissionsHandler, checker))
	r.Delete("/groups/{id}/permissions/{relation}/{user}", wrapHandler(groupRemovePermissionsHandler, checker))

	r.Get("/feeds/", wrapHandler(feedIndexHandler, checker))
	r.Get("/feeds/{id}/permissions", wrapHandler(feedPermissionsHandler, checker))
	r.Get("/feed_versions/", wrapHandler(feedVersionIndexHandler, checker))
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

func handleJson(w http.ResponseWriter, ret any, err error) {
	if err != nil {
		log.Error().Err(err).Msg("admin api error")
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	jj, _ := json.Marshal(ret)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jj)
}

func userIndexHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.ListUsers(r.Context(), auth.ForContext(r.Context()), r.URL.Query().Get("q"))
	handleJson(w, ret, err)
}

func userPermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.User(r.Context(), auth.ForContext(r.Context()), chi.URLParam(r, "id"))
	handleJson(w, ret, err)
}

func tenantIndexHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.ListTenants(r.Context(), auth.ForContext(r.Context()))
	handleJson(w, ret, err)
}

func tenantPermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.TenantPermissions(r.Context(), auth.ForContext(r.Context()), atoi(chi.URLParam(r, "id")))
	handleJson(w, ret, err)
}

func groupIndexHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.ListGroups(r.Context(), auth.ForContext(r.Context()))
	handleJson(w, ret, err)
}

func groupPermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.authn.UserByID(r.Context(), chi.URLParam(r, "id"))
	handleJson(w, ret, err)
}

func groupAddPermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	checkRel, checkRelErr := RelationString(chi.URLParam(r, "relation"))
	if checkRelErr != nil {
		handleJson(w, nil, checkRelErr)
		return
	}
	err := checker.AddGroupPermission(
		r.Context(),
		auth.ForContext(r.Context()),
		chi.URLParam(r, "user"),
		atoi(chi.URLParam(r, "id")),
		checkRel,
	)
	handleJson(w, nil, err)
}

func groupRemovePermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	checkRel, checkRelErr := RelationString(chi.URLParam(r, "relation"))
	if checkRelErr != nil {
		handleJson(w, nil, checkRelErr)
		return
	}
	err := checker.RemoveGroupPermission(
		r.Context(),
		auth.ForContext(r.Context()),
		chi.URLParam(r, "user"),
		atoi(chi.URLParam(r, "id")),
		checkRel,
	)
	handleJson(w, nil, err)
}

func feedIndexHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.ListFeeds(r.Context(), auth.ForContext(r.Context()))
	handleJson(w, ret, err)
}

func feedPermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.FeedPermissions(r.Context(), auth.ForContext(r.Context()), atoi(chi.URLParam(r, "id")))
	handleJson(w, ret, err)
}

func feedVersionIndexHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.ListFeeds(r.Context(), auth.ForContext(r.Context()))
	handleJson(w, ret, err)
}

func feedVersionPermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.FeedVersionPermissions(r.Context(), auth.ForContext(r.Context()), atoi(chi.URLParam(r, "id")))
	handleJson(w, ret, err)
}
