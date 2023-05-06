package authz

import (
	"fmt"
	"os"
	"strings"

	"github.com/interline-io/transitland-lib/tlcsv"
	openfga "github.com/openfga/go-sdk"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type TupleKey struct {
	UserType   ObjectType `json:"user_type"`
	UserName   string     `json:"user_name"`
	ObjectType ObjectType `json:"object_type"`
	ObjectName string     `json:"object_name"`
	Action     Action     `json:"action"`
	Relation   Relation   `json:"relation"`
}

type TestTupleKey struct {
	TupleKey
	Line        int
	Checks      []string
	Test        string
	Expect      string
	Notes       string
	ExpectError bool
	TestAsUser  string
}

func LoadTuples(fn string) ([]TestTupleKey, error) {
	tkeys := []TestTupleKey{}
	if f, err := os.Open(fn); err != nil {
		return nil, err
	} else {
		tlcsv.ReadRows(f, func(row tlcsv.Row) {
			tk := TestTupleKey{}
			tk.Line = row.Line
			tk.UserType = csplit(rowGetString(row, "user")).Type
			tk.UserName = csplit(rowGetString(row, "user")).Name
			tk.ObjectType = csplit(rowGetString(row, "object")).Type
			tk.ObjectName = csplit(rowGetString(row, "object")).Name
			tk.Relation, _ = RelationString(rowGetString(row, "relation"))
			tk.Action, _ = ActionString(rowGetString(row, "action"))
			tk.Checks = strings.Split(rowGetString(row, "check_actions"), " ")
			tk.Test = rowGetString(row, "test")
			tk.Expect = rowGetString(row, "expect")
			tk.Notes = rowGetString(row, "notes")
			tk.TestAsUser = rowGetString(row, "test_as_user")
			if rowGetString(row, "expect_error") == "true" {
				tk.ExpectError = true
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

func (tk TupleKey) String() string {
	r := "relation:" + tk.Relation.String()
	if tk.Action.IsAAction() {
		r = "action:" + tk.Action.String()
	}
	return fmt.Sprintf("%s:%s|%s:%s|%s", tk.UserType.String(), tk.UserName, tk.ObjectType.String(), tk.ObjectName, r)
}

func (tk TupleKey) IsValid() bool {
	if tk.UserName != "" && !tk.UserType.IsAObjectType() {
		return false
	}
	if tk.ObjectName != "" && !tk.ObjectType.IsAObjectType() {
		return false
	}
	if tk.UserName == "" && tk.ObjectName == "" {
		return false
	}
	return tk.Relation.IsARelation() || tk.Action.IsAAction()
}

func (tk TupleKey) WithUser(user string) TupleKey {
	return TupleKey{
		UserType:   UserType,
		UserName:   user,
		ObjectType: tk.ObjectType,
		ObjectName: tk.ObjectName,
		Relation:   tk.Relation,
		Action:     tk.Action,
	}
}

func (tk TupleKey) WithObject(objectType ObjectType, objectName string) TupleKey {
	return TupleKey{
		UserType:   tk.UserType,
		UserName:   tk.UserName,
		ObjectType: objectType,
		ObjectName: objectName,
		Relation:   tk.Relation,
		Action:     tk.Action,
	}

}

func (tk TupleKey) WithRelation(relation Relation) TupleKey {
	return TupleKey{
		UserType:   tk.UserType,
		UserName:   tk.UserName,
		ObjectType: tk.ObjectType,
		ObjectName: tk.ObjectName,
		Relation:   relation,
		Action:     tk.Action,
	}
}

func (tk TupleKey) WithAction(action Action) TupleKey {
	return TupleKey{
		UserType:   tk.UserType,
		UserName:   tk.UserName,
		ObjectType: tk.ObjectType,
		ObjectName: tk.ObjectName,
		Relation:   tk.Relation,
		Action:     action,
	}
}

func fromFGATupleKey(fgatk openfga.TupleKey) TupleKey {
	okeys := csplit(*fgatk.Object)
	ukeys := csplit(*fgatk.User)
	rel, _ := RelationString(*fgatk.Relation)
	act, _ := ActionString(*fgatk.Relation)
	return TupleKey{
		UserType:   ukeys.Type,
		UserName:   ukeys.Name,
		ObjectType: okeys.Type,
		ObjectName: okeys.Name,
		Relation:   rel,
		Action:     act,
	}
}

func (tk TupleKey) FGATupleKey() openfga.TupleKey {
	fgatk := openfga.TupleKey{}
	if tk.UserName != "" {
		fgatk.User = openfga.PtrString(cunsplit(tk.UserType, tk.UserName))
	}
	if tk.ObjectName != "" {
		fgatk.Object = openfga.PtrString(cunsplit(tk.ObjectType, tk.ObjectName))
	}
	if tk.Action.IsAAction() {
		fgatk.Relation = openfga.PtrString(tk.Action.String())
	} else if tk.Relation.IsARelation() {
		fgatk.Relation = openfga.PtrString(tk.Relation.String())
	}
	return fgatk
}

type cs struct {
	Type ObjectType
	Name string
}

func csplit(v string) cs {
	a := strings.Split(v, ":")
	ret := cs{}
	if len(a) > 1 {
		ret.Type, _ = ObjectTypeString(a[0])
		ret.Name = a[1]
	} else if len(a) > 0 {
		ret.Type, _ = ObjectTypeString(a[0])
	}
	return ret
}

func cunsplit(a ObjectType, b string) string {
	if b == "" {
		return a.String()
	}
	return fmt.Sprintf("%s:%s", a.String(), b)
}
