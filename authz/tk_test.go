package authz

import (
	"fmt"
	"os"
	"strings"

	"github.com/interline-io/transitland-lib/tlcsv"
)

type TestTupleKey struct {
	TupleKey
	TestName          string
	Line              int
	Checks            []string
	Test              string
	Expect            string
	Notes             string
	ExpectError       bool
	CheckAsUser       string
	ExpectErrorAsUser bool
}

func (tk TestTupleKey) String() string {
	return fmt.Sprintf("line:%d|%s", tk.Line, tk.TupleKey.String())
}

func LoadTuples(fn string) ([]TestTupleKey, error) {
	tkeys := []TestTupleKey{}
	if f, err := os.Open(fn); err != nil {
		return nil, err
	} else {
		tlcsv.ReadRows(f, func(row tlcsv.Row) {
			tk := TestTupleKey{}
			tk.Line = row.Line
			tk.Subject = NewEntityKeySplit(rowGetString(row, "user"))
			tk.Object = NewEntityKeySplit(rowGetString(row, "object"))
			tk.Relation, _ = RelationString(rowGetString(row, "relation"))
			tk.Action, _ = ActionString(rowGetString(row, "action"))
			tk.Checks = strings.Split(rowGetString(row, "check_actions"), " ")
			tk.Test = rowGetString(row, "test")
			tk.Expect = rowGetString(row, "expect")
			tk.TestName = rowGetString(row, "test_name")
			tk.Notes = rowGetString(row, "notes")
			tk.CheckAsUser = rowGetString(row, "check_as_user")
			if rowGetString(row, "expect_error") == "true" {
				tk.ExpectError = true
			}
			if rowGetString(row, "expect_error_as_user") == "true" {
				tk.ExpectErrorAsUser = true
			}

			tkeys = append(tkeys, tk)
		})
	}
	return tkeys, nil
}

func rowGetString(row tlcsv.Row, col string) string {
	a, _ := row.Get(col)
	return a
}
