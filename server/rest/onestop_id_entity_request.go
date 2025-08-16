package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/interline-io/transitland-server/server/model"
)

// Query returns a GraphQL query string and variables.
func onestopIdEntityRedirectHandler(graphqlHandler http.Handler, w http.ResponseWriter, r *http.Request) {
	cfg := model.ForContext(r.Context())
	onestop_id := chi.URLParam(r, "onestop_id")
	fmt.Println("onestop_id:", onestop_id)
	var redirectUrl string
	if strings.HasPrefix(onestop_id, "f-") {
		redirectUrl = fmt.Sprintf("%s/feeds/%s", cfg.RestPrefix, onestop_id)
		// redirect to feeds/
	} else if strings.HasPrefix(onestop_id, "o-") {
		redirectUrl = fmt.Sprintf("%s/operators/%s", cfg.RestPrefix, onestop_id)
	} else if strings.HasPrefix(onestop_id, "s-") {
		redirectUrl = fmt.Sprintf("%s/stops/%s", cfg.RestPrefix, onestop_id)
	} else if strings.HasPrefix(onestop_id, "r-") {
		redirectUrl = fmt.Sprintf("%s/routes/%s", cfg.RestPrefix, onestop_id)
	}
	if redirectUrl != "" {
		w.Header().Add("Location", redirectUrl)
		w.WriteHeader(http.StatusFound)
	} else {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
}
