package authz

import (
	"context"
	"strings"
	"testing"
)

func TestFGA(t *testing.T) {
	cc, err := NewFGAClient("", "http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	modelId, err := createTestStoreAndModel(cc, "test", "../test/authz/test.model", true)
	if err != nil {
		t.Fatal(err)
	}
	cc.Model = modelId

	// Add test tuples
	tkeys, err := LoadTuples("../test/authz/test.csv")
	if err != nil {
		t.Fatal(err)
	}
	for _, tk := range tkeys {
		if err := cc.WriteTuple(context.Background(), tk); err != nil {
			t.Fatal(err)
		}
	}

	// Test assertions
	checks, err := LoadTuples("../test/authz/assert.csv")
	if err != nil {
		t.Fatal(err)
	}
	for _, tk := range checks {
		ok, err := cc.Check(context.Background(), tk)
		if err != nil {
			t.Fatal(err)
		}
		if ok != tk.Assert {
			t.Errorf("%s/%s/%s got %t, expected %t", tk.User, tk.Object, tk.Relation, ok, tk.Assert)
		}
	}

	// List objects
	checkusers := []string{"admin", "ian", "drew", "kapeel"}
	rels := []string{"can_view", "can_edit"}
	for _, user := range checkusers {
		for _, rel := range rels {
			if objs, err := cc.ListObjects(context.Background(), TupleKey{User: "user:" + user, Object: "feed", Relation: rel}); err != nil {
				t.Fatal(err)
			} else {
				t.Logf("user %s: %s: %s", user, rel, strings.Join(objs, " "))
			}
		}
	}
}
