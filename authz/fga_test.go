package authz

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFGAClient(t *testing.T) {
	if os.Getenv("TL_TEST_FGA_ENDPOINT") == "" {
		t.Skip("no TL_TEST_FGA_ENDPOINT set, skipping")
		return
	}

	// Test assertions
	checks, err := LoadTuples("../test/authz/tls.csv")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("GetObjectTuples", func(t *testing.T) {
		fgac := newTestFGAClient(t)
		for _, tk := range checks {
			if tk.Test != "get" {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				tks, err := fgac.GetObjectTuples(context.Background(), tk.TupleKey)
				if err != nil {
					t.Error(err)
				}
				expect := strings.Split(tk.Expect, " ")
				var got []string
				for _, v := range tks {
					got = append(got, fmt.Sprintf("%s:%s:%s", v.Subject.Type, v.Subject.Name, v.Relation))
				}
				assert.ElementsMatch(t, expect, got, "usertype:username:relation does not match")

			})
		}
	})

	t.Run("Check", func(t *testing.T) {
		fgac := newTestFGAClient(t)
		for _, tk := range checks {
			if tk.Test != "check" {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				for _, checkAction := range tk.Checks {
					expect := true
					if strings.HasPrefix(checkAction, "+") {
						checkAction = strings.TrimPrefix(checkAction, "+")
					} else if strings.HasPrefix(checkAction, "-") {
						expect = false
						checkAction = strings.TrimPrefix(checkAction, "-")
					}
					var err error
					tk := tk
					tk.TupleKey.Action, err = ActionString(checkAction)
					if err != nil {
						t.Fatal(err)
					}
					ok, err := fgac.Check(context.Background(), tk.TupleKey)
					if err != nil {
						t.Fatal(err)
					}
					if ok && !expect {
						t.Errorf("got %t, expected %t", ok, expect)
					}
					if !ok && expect {
						t.Errorf("got %t, expected %t", ok, expect)
					}
				}
			})
		}
	})

	t.Run("ListObjects", func(t *testing.T) {
		fgac := newTestFGAClient(t)
		for _, tk := range checks {
			if tk.Test != "list" {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				for _, checkAction := range tk.Checks {
					tk := tk
					tk.TupleKey.Action, err = ActionString(checkAction)
					if err != nil {
						t.Fatal(err)
					}
					objs, err := fgac.ListObjects(context.Background(), tk.TupleKey)
					if err != nil {
						t.Fatal(err)
					}
					var gotIds []int
					for _, v := range objs {
						gotIds = append(gotIds, v.Object.ID())
					}
					expIds := mapStrInt(tk.Expect)
					assert.ElementsMatch(t, expIds, gotIds, "object ids")
				}
			})
		}
	})

	t.Run("WriteTuple", func(t *testing.T) {
		fgac := newTestFGAClient(t)
		for _, tk := range checks {
			if tk.Test != "write" {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				// Write tuple and check if error was expected
				err := fgac.WriteTuple(context.Background(), tk.TupleKey)
				if !checkExpectError(t, err, tk.ExpectError) {
					return
				}
				// Check was written
				tks, err := fgac.GetObjectTuples(context.Background(), tk.TupleKey)
				if err != nil {
					t.Error(err)
				}
				var gotTks []string
				for _, v := range tks {
					gotTks = append(gotTks, fmt.Sprintf("%s:%s:%s", v.Subject.Type, v.Subject.Name, v.Relation))
				}
				checkTk := fmt.Sprintf("%s:%s:%s", tk.Subject.Type, tk.Subject.Name, tk.Relation)
				assert.Contains(t, gotTks, checkTk, "written tuple not found in updated object tuples")
			})
		}
	})

	t.Run("DeleteTuple", func(t *testing.T) {
		fgac := newTestFGAClient(t)
		for _, tk := range checks {
			if tk.Test != "delete" {
				continue
			}
			t.Run(tk.String(), func(t *testing.T) {
				err := fgac.DeleteTuple(context.Background(), tk.TupleKey)
				if !checkExpectError(t, err, tk.ExpectError) {
					return
				}
			})
		}
	})
}

func newTestFGAClient(t testing.TB) *FGAClient {
	cfg := newTestConfig()
	fgac, err := NewFGAClient(cfg.FGAEndpoint, cfg.FGAStoreID, cfg.FGAModelID)
	if err != nil {
		t.Fatal(err)
		return nil
	}
	if cfg.FGALoadModelFile != "" {
		if _, err := fgac.CreateStore(context.Background(), "test"); err != nil {
			t.Fatal(err)
		}
		if _, err := fgac.CreateModel(context.Background(), cfg.FGALoadModelFile); err != nil {
			t.Fatal(err)
		}
	}
	if cfg.FGALoadTupleFile != "" {
		tkeys, err := LoadTuples(cfg.FGALoadTupleFile)
		if err != nil {
			t.Fatal(err)
			return nil
		}
		for _, tk := range tkeys {
			if tk.Test != "" {
				continue
			}
			if !tk.Relation.IsARelation() {
				continue
			}
			if err := fgac.WriteTuple(context.Background(), tk.TupleKey); err != nil {
				t.Fatal(err)
				return nil
			}
		}
	}
	return fgac
}

func checkExpectError(t testing.TB, err error, expect bool) bool {
	if err != nil && !expect {
		t.Errorf("got error '%s', did not expect error", err.Error())
		return false
	}
	if err == nil && expect {
		t.Errorf("got no error, expected error")
		return false
	}
	if err != nil {
		return false
	}
	return true
}
