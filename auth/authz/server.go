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
)

func NewServer(checker *Checker) (http.Handler, error) {
	router := chi.NewRouter()

	/////////////////
	// USERS
	/////////////////

	router.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.UserList(r.Context(), &UserListRequest{Q: r.URL.Query().Get("q")})
		handleJson(w, ret, err)
	})
	router.Get("/users/{user_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.User(r.Context(), &UserRequest{Id: chi.URLParam(r, "user_id")})
		handleJson(w, ret, err)
	})

	/////////////////
	// TENANTS
	/////////////////

	router.Get("/tenants", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.TenantList(r.Context(), &TenantListRequest{})
		handleJson(w, ret, err)
	})
	router.Get("/tenants/{tenant_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.TenantPermissions(r.Context(), &TenantRequest{Id: checkId(r, "tenant_id")})
		handleJson(w, ret, err)
	})
	router.Post("/tenants/{tenant_id}", func(w http.ResponseWriter, r *http.Request) {
		check := Tenant{}
		if err := parseJson(r.Body, &check); err != nil {
			handleJson(w, nil, err)
			return
		}
		check.Id = checkId(r, "tenant_id")
		_, err := checker.TenantSave(r.Context(), &TenantSaveRequest{Tenant: &check})
		handleJson(w, nil, err)
	})
	router.Post("/tenants/{tenant_id}/groups", func(w http.ResponseWriter, r *http.Request) {
		check := Group{}
		if err := parseJson(r.Body, &check); err != nil {
			handleJson(w, nil, err)
			return
		}
		_, err := checker.TenantCreateGroup(r.Context(), &TenantCreateGroupRequest{Id: checkId(r, "tenant_id"), Group: &check})
		handleJson(w, nil, err)
	})
	router.Post("/tenants/{tenant_id}/permissions/{relation}/{user}", func(w http.ResponseWriter, r *http.Request) {
		entId, userRel, err := checkUserRel(r, "tenant_id")
		if err != nil {
			handleJson(w, nil, err)
			return
		}
		_, err = checker.TenantAddPermission(r.Context(), &TenantModifyPermissionRequest{Id: entId, UserRelation: userRel})
		handleJson(w, nil, err)
	})
	router.Delete("/tenants/{tenant_id}/permissions/{relation}/{user}", func(w http.ResponseWriter, r *http.Request) {
		entId, userRel, err := checkUserRel(r, "tenant_id")
		if err != nil {
			handleJson(w, nil, err)
			return
		}
		_, err = checker.TenantRemovePermission(r.Context(), &TenantModifyPermissionRequest{Id: entId, UserRelation: userRel})
		handleJson(w, nil, err)
	})

	/////////////////
	// GROUPS
	/////////////////

	router.Get("/groups", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.GroupList(r.Context(), &GroupListRequest{})
		handleJson(w, ret, err)
	})
	router.Post("/groups/{group_id}", func(w http.ResponseWriter, r *http.Request) {
		check := Group{}
		if err := parseJson(r.Body, &check); err != nil {
			handleJson(w, nil, err)
			return
		}
		check.Id = checkId(r, "group_id")
		_, err := checker.GroupSave(r.Context(), &GroupSaveRequest{Group: &check})
		handleJson(w, nil, err)
	})
	router.Get("/groups/{group_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.GroupPermissions(r.Context(), &GroupRequest{Id: checkId(r, "group_id")})
		handleJson(w, ret, err)
	})
	router.Post("/groups/{group_id}/permissions/{relation}/{user}", func(w http.ResponseWriter, r *http.Request) {
		entId, userRel, err := checkUserRel(r, "group_id")
		if err != nil {
			handleJson(w, nil, err)
			return
		}
		_, err = checker.GroupAddPermission(r.Context(), &GroupModifyPermissionRequest{Id: entId, UserRelation: userRel})
		handleJson(w, nil, err)
	})
	router.Delete("/groups/{group_id}/permissions/{relation}/{user}", func(w http.ResponseWriter, r *http.Request) {
		entId, userRel, err := checkUserRel(r, "group_id")
		if err != nil {
			handleJson(w, nil, err)
			return
		}
		_, err = checker.GroupRemovePermission(r.Context(), &GroupModifyPermissionRequest{Id: entId, UserRelation: userRel})
		handleJson(w, nil, err)
	})

	/////////////////
	// FEEDS
	/////////////////

	router.Get("/feeds", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.FeedList(r.Context(), &FeedListRequest{})
		handleJson(w, ret, err)
	})
	router.Get("/feeds/{feed_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.FeedPermissions(r.Context(), &FeedRequest{Id: checkId(r, "feed_id")})
		handleJson(w, ret, err)
	})
	router.Post("/feeds/{feed_id}/group", func(w http.ResponseWriter, r *http.Request) {
		check := FeedSetGroupRequest{}
		if err := parseJson(r.Body, &check); err != nil {
			handleJson(w, nil, err)
			return
		}
		check.Id = checkId(r, "feed_id")
		_, err := checker.FeedSetGroup(r.Context(), &check)
		handleJson(w, nil, err)
	})

	/////////////////
	// FEED VERSIONS
	/////////////////

	router.Get("/feed_versions", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.FeedVersionList(r.Context(), &FeedVersionListRequest{})
		handleJson(w, ret, err)
	})
	router.Get("/feed_versions/{feed_version_id}", func(w http.ResponseWriter, r *http.Request) {
		ret, err := checker.FeedVersionPermissions(r.Context(), &FeedVersionRequest{Id: checkId(r, "feed_version_id")})
		handleJson(w, ret, err)
	})
	router.Post("/feed_versions/{feed_version_id}/permissions/{relation}/{user}", func(w http.ResponseWriter, r *http.Request) {
		entId, userRel, err := checkUserRel(r, "feed_version_id")
		if err != nil {
			handleJson(w, nil, err)
			return
		}
		_, err = checker.FeedVersionAddPermission(r.Context(), &FeedVersionModifyPermissionRequest{Id: entId, UserRelation: userRel})
		handleJson(w, nil, err)
	})
	router.Delete("/feed_versions/{feed_version_id}/permissions/{relation}/{user}", func(w http.ResponseWriter, r *http.Request) {
		entId, userRel, err := checkUserRel(r, "feed_version_id")
		if err != nil {
			handleJson(w, nil, err)
			return
		}
		_, err = checker.FeedVersionRemovePermission(r.Context(), &FeedVersionModifyPermissionRequest{Id: entId, UserRelation: userRel})
		handleJson(w, nil, err)
	})

	return router, nil
}

func handleJson(w http.ResponseWriter, ret any, err error) {
	if err == ErrUnauthorized {
		log.Error().Err(err).Msg("unauthorized")
		http.Error(w, "error", http.StatusUnauthorized)
		return
	} else if err != nil {
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

func checkId(r *http.Request, key string) int64 {
	v, _ := strconv.Atoi(chi.URLParam(r, key))
	return int64(v)
}

func parseJson(r io.Reader, v any) error {
	data, err := ioutil.ReadAll(io.LimitReader(r, 1_000_000))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func checkUserRel(r *http.Request, idKey string) (int64, *UserRelation, error) {
	id := int64(0)
	tk := &UserRelation{}
	var err error
	if tk.Relation, err = RelationString(chi.URLParam(r, "relation")); err != nil {
		return 0, tk, err
	}
	if vid, err := strconv.Atoi(chi.URLParam(r, idKey)); err != nil {
		return 0, tk, errors.New("invalid id")
	} else {
		id = int64(vid)
	}
	tk.UserId = chi.URLParam(r, "user")
	return id, tk, nil
}
