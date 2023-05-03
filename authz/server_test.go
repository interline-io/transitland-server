package authz

import (
	"testing"

	"github.com/interline-io/transitland-server/internal/testfinder"
)

func TestServer(t *testing.T) {
	te := testfinder.Finders(t, nil, nil)
	checker, err := newTestChecker(t, AuthzConfig{}, te.Finder)
	if err != nil {
		t.Fatal(err)
	}
	srv, err := NewServer(checker)
	if err != nil {
		t.Fatal(err)
	}
	_ = srv
}
