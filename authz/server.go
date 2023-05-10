package authz

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/interline-io/transitland-lib/log"
	"github.com/interline-io/transitland-server/auth"
)

func NewServer(checker *Checker) (http.Handler, error) {
	r := chi.NewRouter()
	r.Get("/users", wrapHandler(userIndexHandler, checker))
	r.Get("/users/{user_id}", wrapHandler(userPermissionsHandler, checker))

	r.Get("/tenants", wrapHandler(tenantIndexHandler, checker))
	r.Get("/tenants/{tenant_id}", wrapHandler(tenantPermissionsHandler, checker))
	r.Post("/tenants/{tenant_id}/groups", wrapHandler(tenantCreateGroupHandler, checker))

	r.Get("/groups", wrapHandler(groupIndexHandler, checker))
	r.Post("/groups/{group_id}", wrapHandler(groupSave, checker))
	r.Get("/groups/{group_id}", wrapHandler(groupPermissionsHandler, checker))
	r.Post("/groups/{group_id}/permissions/{relation}/{user}", wrapHandler(groupAddPermissionsHandler, checker))
	r.Delete("/groups/{group_id}/permissions/{relation}/{user}", wrapHandler(groupRemovePermissionsHandler, checker))

	r.Get("/feeds/", wrapHandler(feedIndexHandler, checker))
	r.Get("/feeds/{feed_id}", wrapHandler(feedPermissionsHandler, checker))
	r.Post("/feeds/{feed_id}/group", wrapHandler(feedSetGroupHandler, checker))

	r.Get("/feed_versions/", wrapHandler(feedVersionIndexHandler, checker))
	r.Get("/feed_versions/{feed_version_id}", wrapHandler(feedVersionPermissionsHandler, checker))
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
	if ret == nil {
		ret = map[string]bool{"success": true}
	}
	jj, _ := json.Marshal(ret)
	w.Header().Add("Content-Type", "application/json")
	w.Write(jj)
}

////////////
// Users
////////////

func userIndexHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.UserList(r.Context(), auth.ForContext(r.Context()), r.URL.Query().Get("q"))
	handleJson(w, ret, err)
}

func userPermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.User(r.Context(), auth.ForContext(r.Context()), chi.URLParam(r, "user_id"))
	handleJson(w, ret, err)
}

////////////
// Tenants
////////////

func tenantIndexHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.TenantList(r.Context(), auth.ForContext(r.Context()))
	handleJson(w, ret, err)
}

func tenantPermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.TenantPermissions(r.Context(), auth.ForContext(r.Context()), checkId(r, "tenant_id"))
	handleJson(w, ret, err)
}

func tenantCreateGroupHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	checkReq := GroupResponse{}
	if err := parseJson(r.Body, &checkReq); err != nil {
		handleJson(w, nil, err)
		return
	}
	newId, err := checker.TenantCreateGroup(r.Context(), auth.ForContext(r.Context()), checkId(r, "tenant_id"), checkReq.Name)
	_ = newId
	handleJson(w, nil, err)
}

func groupSave(w http.ResponseWriter, r *http.Request, checker *Checker) {
	checkReq := GroupResponse{}
	if err := parseJson(r.Body, &checkReq); err != nil {
		handleJson(w, nil, err)
		return
	}
	_, err := checker.GroupSave(r.Context(), auth.ForContext(r.Context()), checkId(r, "group_id"), checkReq.Name)
	handleJson(w, nil, err)
}

////////////
// Groups
////////////

func groupIndexHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.GroupList(r.Context(), auth.ForContext(r.Context()))
	handleJson(w, ret, err)
}

func groupPermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.GroupPermissions(r.Context(), auth.ForContext(r.Context()), checkId(r, "group_id"))
	handleJson(w, ret, err)
}

func groupAddPermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	checkRel, err := checkRelParams(r, "group_id")
	if err != nil {
		handleJson(w, nil, err)
		return
	}
	err = checker.GroupAddPermission(r.Context(), checkRel.User, checkRel.RelUser, checkRel.ID, checkRel.Relation)
	handleJson(w, nil, err)
}

func groupRemovePermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	checkRel, err := checkRelParams(r, "group_id")
	if err != nil {
		handleJson(w, nil, err)
		return
	}
	err = checker.GroupRemovePermission(r.Context(), checkRel.User, checkRel.RelUser, checkRel.ID, checkRel.Relation)
	handleJson(w, nil, err)
}

func feedIndexHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.ListFeeds(r.Context(), auth.ForContext(r.Context()))
	handleJson(w, ret, err)
}

////////////
// Feeds
////////////

func feedSetGroupHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	checkParams := struct {
		GroupID int `json:"group_id"`
	}{}
	if err := parseJson(r.Body, &checkParams); err != nil {
		handleJson(w, nil, err)
		return
	}
	err := checker.FeedSetGroup(r.Context(), auth.ForContext(r.Context()), checkId(r, "feed_version_id"), checkParams.GroupID)
	handleJson(w, nil, err)
}

func feedPermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.FeedPermissions(r.Context(), auth.ForContext(r.Context()), checkId(r, "feed_id"))
	handleJson(w, ret, err)
}

////////////
// Feed Versions
////////////

func feedVersionIndexHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.ListFeeds(r.Context(), auth.ForContext(r.Context()))
	handleJson(w, ret, err)
}

func feedVersionPermissionsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	ret, err := checker.FeedVersionPermissions(r.Context(), auth.ForContext(r.Context()), checkId(r, "feed_version_id"))
	handleJson(w, ret, err)
}

////////////
// Utility
////////////

type checkRel struct {
	User     auth.User
	RelUser  string
	Relation Relation
	ID       int
}

func checkId(r *http.Request, key string) int {
	return atoi(chi.URLParam(r, key))
}

func checkRelParams(r *http.Request, idKey string) (checkRel, error) {
	tk := checkRel{}
	var err error
	if tk.User = auth.ForContext(r.Context()); tk.User == nil {
		return tk, errors.New("unauthorized")
	}
	if tk.Relation, err = RelationString(chi.URLParam(r, "relation")); err != nil {
		return tk, err
	}
	if tk.ID, err = strconv.Atoi(chi.URLParam(r, idKey)); err != nil {
		return tk, errors.New("invalid id")
	}
	tk.RelUser = chi.URLParam(r, "user")
	return tk, nil
}

func parseJson(r io.Reader, v any) error {
	data, err := ioutil.ReadAll(io.LimitReader(r, 1_000_000))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
