package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/model"
)

type AuthzConfig struct {
	Auth0Domain       string
	Auth0ClientID     string
	Auth0ClientSecret string
	FGAModelID        string
	FGAEndpoint       string
	FGATestModelPath  string
	FGATestTuplesPath string
}

func NewServer(finder model.Finder) (http.Handler, error) {
	cfg := AuthzConfig{
		Auth0Domain:       os.Getenv("TL_AUTH0_DOMAIN"),
		Auth0ClientID:     os.Getenv("TL_AUTH0_CLIENT_ID"),
		Auth0ClientSecret: os.Getenv("TL_AUTH0_CLIENT_SECRET"),
		FGAModelID:        os.Getenv("TL_FGA_MODEL_ID"),
		FGAEndpoint:       os.Getenv("TL_FGA_ENDPOINT"),
		FGATestModelPath:  os.Getenv("TL_FGA_TEST_MODEL_PATH"),
		FGATestTuplesPath: os.Getenv("TL_FGA_TEST_TUPLES_PATH"),
	}
	checker, err := checkerFromConfig(cfg, finder)
	if err != nil {
		return nil, err
	}
	r := chi.NewRouter()
	r.Get("/users", wrapHandler(usersHandler, checker))
	r.Get("/users/{id}", wrapHandler(userHandler, checker))
	r.Get("/groups", groupsHandler)
	r.Get("/groups/{id}", groupHandler)
	r.Get("/objects", wrapHandler(indexObjectsHandler, checker))
	r.Get("/objects/{id}", getObjectHandler)
	r.Put("/objects/{id}/tuples", getObjectHandler)
	r.Delete("/objects/{id}/tuples", getObjectHandler)
	return r, nil
}

func checkerFromConfig(cfg AuthzConfig, finder model.Finder) (*Checker, error) {
	auth0c, err := NewAuth0Client(cfg.Auth0Domain, cfg.Auth0ClientID, cfg.Auth0ClientSecret)
	if err != nil {
		return nil, err
	}
	fgac, err := NewFGAClient(cfg.FGAModelID, cfg.FGAEndpoint)
	if err != nil {
		return nil, err
	}
	if cfg.FGATestModelPath != "" {
		modelId, err := createTestStoreAndModel(fgac, "test", cfg.FGATestModelPath, true)
		if err != nil {
			return nil, err
		}
		fgac.Model = modelId
	}
	if cfg.FGATestTuplesPath != "" {
		tkeys, err := LoadTuples(cfg.FGATestTuplesPath)
		if err != nil {
			return nil, err
		}
		for _, tk := range tkeys {
			if err := fgac.WriteTuple(context.Background(), tk); err != nil {
				return nil, err
			}
		}

	}
	checker := NewChecker(auth0c, fgac, finder, nil)
	return checker, err
}

func wrapHandler(next func(http.ResponseWriter, *http.Request, *Checker), checker *Checker) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next(w, r, checker)
	})
}

func usersHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	users, err := checker.authn.Users(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	jj, _ := json.Marshal(users)
	w.Write(jj)
}

func userHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	user, err := checker.authn.UserByID(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	jj, _ := json.Marshal(user)
	w.Write(jj)
}

func groupsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("groups:")
}

func groupHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("group:")
}

func indexObjectsHandler(w http.ResponseWriter, r *http.Request, checker *Checker) {
	u := auth.ForContext(r.Context())
	if u == nil {
		http.Error(w, "not logged in", http.StatusUnauthorized)
	}
	feeds, err := checker.Feeds(r.Context(), u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jj, _ := json.Marshal(feeds)
	w.Write(jj)
}

func getObjectHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("object:")
}
