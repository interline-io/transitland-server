package authz

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

func NewServer() (http.Handler, error) {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})
	r.HandleFunc("/users", usersHandler)
	return r, nil
	// r := chi.NewRouter()
	// r.HandleFunc("/", usersHandler)
	// r.HandleFunc("/users", usersHandler)
	// r.Get("/groups", groupsHandler)
	// r.Get("/groups/{id}", groupHandler)
	// r.Get("/objects", indexObjectsHandler)
	// r.Get("/objects/{id}", getObjectHandler)
	// r.Put("/objects/{id}/tuples", getObjectHandler)
	// r.Delete("/objects/{id}/tuples", getObjectHandler)
	// return r, nil
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("users:")
	w.Write([]byte("hi"))
}

func groupsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("groups:")
}

func groupHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("group:")
}

func indexObjectsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("objects:")
}

func getObjectHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("object:")
}
