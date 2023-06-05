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
	router := chi.NewRouter()

	/////////////////
	// USERS
	/////////////////

	router.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.UserList(r.Context(), checkUser(r), r.URL.Query().Get("q"))
		handleJson(w, ret, err)
	})
	router.Get("/users/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.User(r.Context(), checkUser(r), chi.URLParam(r, "user_id"))
		handleJson(w, ret, err)
	})

	/////////////////
	// TENANTS
	/////////////////

	router.Get("/tenants", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.TenantList(r.Context(), checkUser(r))
		handleJson(w, ret, err)
	})
	router.Get("/tenants/{tenant_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.TenantPermissions(r.Context(), checkUser(r), checkId(r, "tenant_id"))
		handleJson(w, ret, err)
	})
	router.Post("/tenants/{tenant_id}", func(w http.ResponseWriter, r *http.Request) {
		check := struct {
			Name string `json:"name"`
		}{}
		if err := parseJson(r.Body, &check); err != nil {
			handleJson(w, nil, err)
			return
		}
		_, err := checker.TenantSave(r.Context(), checkUser(r), checkId(r, "tenant_id"), check.Name)
		handleJson(w, nil, err)
	})
	router.Post("/tenants/{tenant_id}/groups", func(w http.ResponseWriter, r *http.Request) {
		check := struct {
			Name string `json:"name"`
		}{}
		if err := parseJson(r.Body, &check); err != nil {
			handleJson(w, nil, err)
			return
		}
		newId, err := checker.TenantCreateGroup(r.Context(), checkUser(r), checkId(r, "tenant_id"), check.Name)
		_ = newId
		handleJson(w, nil, err)
	})
	router.Post("/tenants/{tenant_id}/permissions/{relation}/{user}", func(w http.ResponseWriter, r *http.Request) {
		checkRel, err := checkRelParams(r, "tenant_id")
		if err != nil {
			handleJson(w, nil, err)
			return
		}
		err = checker.TenantAddPermission(r.Context(), checkRel.User, checkRel.ID, checkRel.RelUser, checkRel.Relation)
		handleJson(w, nil, err)
	})
	router.Delete("/tenants/{tenant_id}/permissions/{relation}/{user}", func(w http.ResponseWriter, r *http.Request) {
		checkRel, err := checkRelParams(r, "tenant_id")
		if err != nil {
			handleJson(w, nil, err)
			return
		}
		err = checker.TenantRemovePermission(r.Context(), checkRel.User, checkRel.ID, checkRel.RelUser, checkRel.Relation)
		handleJson(w, nil, err)
	})

	/////////////////
	// GROUPS
	/////////////////

	router.Get("/groups", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.GroupList(r.Context(), checkUser(r))
		handleJson(w, ret, err)
	})
	router.Post("/groups/{group_id}", func(w http.ResponseWriter, r *http.Request) {
		check := struct {
			Name string `json:"name"`
		}{}
		if err := parseJson(r.Body, &check); err != nil {
			handleJson(w, nil, err)
			return
		}
		_, err := checker.GroupSave(r.Context(), checkUser(r), checkId(r, "group_id"), check.Name)
		handleJson(w, nil, err)
	})
	router.Get("/groups/{group_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.GroupPermissions(r.Context(), checkUser(r), checkId(r, "group_id"))
		handleJson(w, ret, err)
	})
	router.Post("/groups/{group_id}/permissions/{relation}/{user}", func(w http.ResponseWriter, r *http.Request) {
		checkRel, err := checkRelParams(r, "group_id")
		if err != nil {
			handleJson(w, nil, err)
			return
		}
		err = checker.GroupAddPermission(r.Context(), checkRel.User, checkRel.RelUser, checkRel.ID, checkRel.Relation)
		handleJson(w, nil, err)
	})
	router.Delete("/groups/{group_id}/permissions/{relation}/{user}", func(w http.ResponseWriter, r *http.Request) {
		checkRel, err := checkRelParams(r, "group_id")
		if err != nil {
			handleJson(w, nil, err)
			return
		}
		err = checker.GroupRemovePermission(r.Context(), checkRel.User, checkRel.RelUser, checkRel.ID, checkRel.Relation)
		handleJson(w, nil, err)
	})

	/////////////////
	// FEEDS
	/////////////////

	router.Get("/feeds", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.FeedList(r.Context(), checkUser(r))
		handleJson(w, ret, err)
	})
	router.Get("/feeds/{feed_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.FeedPermissions(r.Context(), checkUser(r), checkId(r, "feed_id"))
		handleJson(w, ret, err)
	})
	router.Post("/feeds/{feed_id}/group", func(w http.ResponseWriter, r *http.Request) {
		checkParams := struct {
			GroupID int `json:"group_id"`
		}{}
		if err := parseJson(r.Body, &checkParams); err != nil {
			handleJson(w, nil, err)
			return
		}
		err := checker.FeedSetGroup(r.Context(), checkUser(r), checkId(r, "feed_version_id"), checkParams.GroupID)
		handleJson(w, nil, err)
	})

	/////////////////
	// FEED VERSIONS
	/////////////////

	router.Get("/feed_versions", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.FeedVersionList(r.Context(), checkUser(r))
		handleJson(w, ret, err)
	})
	router.Get("/feed_versions/{feed_version_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.FeedVersionPermissions(r.Context(), checkUser(r), checkId(r, "feed_version_id"))
		handleJson(w, ret, err)
	})
	router.Post("/feed_versions/{feed_version_id}/permissions/{relation}/{user}", func(w http.ResponseWriter, r *http.Request) {
		checkRel, err := checkRelParams(r, "feed_version_id")
		if err != nil {
			handleJson(w, nil, err)
			return
		}
		err = checker.FeedVersionAddPermission(r.Context(), checkRel.User, checkRel.RelUser, checkRel.ID, checkRel.Relation)
		handleJson(w, nil, err)
	})
	router.Delete("/feed_versions/{feed_version_id}/permissions/{relation}/{user}", func(w http.ResponseWriter, r *http.Request) {
		checkRel, err := checkRelParams(r, "feed_version_id")
		if err != nil {
			handleJson(w, nil, err)
			return
		}
		err = checker.FeedVersionRemovePermission(r.Context(), checkRel.User, checkRel.RelUser, checkRel.ID, checkRel.Relation)
		handleJson(w, nil, err)
	})

	return router, nil
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

func checkUser(r *http.Request) auth.User {
	return auth.ForContext(r.Context())
}

func checkId(r *http.Request, key string) int {
	v, _ := strconv.Atoi(chi.URLParam(r, key))
	return v
}

func parseJson(r io.Reader, v any) error {
	data, err := ioutil.ReadAll(io.LimitReader(r, 1_000_000))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

type checkRel struct {
	User     auth.User
	RelUser  string
	Relation Relation
	ID       int
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
