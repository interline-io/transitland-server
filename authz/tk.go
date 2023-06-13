package authz

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/interline-io/transitland-lib/tlcsv"
	openfga "github.com/openfga/go-sdk"
)

type EntityKey struct {
	Type ObjectType `json:"Type"`
	Name string     `json:"Name"`
}

func NewEntityKey(t ObjectType, name string) EntityKey {
	return EntityKey{Type: t, Name: name}
}

func NewEntityKeySplit(v string) EntityKey {
	ret := EntityKey{}
	a := strings.Split(v, ":")
	if len(a) > 1 {
		ret.Type, _ = ObjectTypeString(a[0])
		ret.Name = a[1]
	} else if len(a) > 0 {
		ret.Type, _ = ObjectTypeString(a[0])
	}
	return ret
}

func NewEntityID(t ObjectType, id int) EntityKey {
	return EntityKey{Type: t, Name: strconv.Itoa(id)}
}

func (ek EntityKey) ID() int {
	v, _ := strconv.Atoi(ek.Name)
	return v
}

func (ek EntityKey) String() string {
	if ek.Name == "" {
		return ek.Type.String()
	}
	return fmt.Sprintf("%s:%s", ek.Type.String(), ek.Name)

}

type TupleKey struct {
	Subject  EntityKey
	Object   EntityKey
	Action   Action   `json:"action"`
	Relation Relation `json:"relation"`
}

func NewTupleKey() TupleKey { return TupleKey{} }

func (tk TupleKey) String() string {
	r := ""
	if tk.Relation.IsARelation() {
		r = "|relation:" + tk.Relation.String()
	} else if tk.Action.IsAAction() {
		r = "|action:" + tk.Action.String()
	}
	return fmt.Sprintf("%s|%s%s", tk.Subject.String(), tk.Object.String(), r)
}

func (tk TupleKey) IsValid() bool {
	return tk.Validate() == nil
}

func (tk TupleKey) Validate() error {
	if tk.Subject.Name != "" && !tk.Subject.Type.IsAObjectType() {
		return errors.New("invalid user type")
	}
	if tk.Object.Name != "" && !tk.Object.Type.IsAObjectType() {
		return errors.New("invalid object type")
	}
	if tk.Subject.Name == "" && tk.Object.Name == "" {
		return errors.New("user name or object name is required")
	}
	if tk.Subject.Name != "" && tk.Object.Name != "" {
		if tk.Action == 0 && !tk.Relation.IsARelation() {
			return errors.New("invalid relation")
		}
		if tk.Relation == 0 && !tk.Action.IsAAction() {
			return errors.New("invalid action")
		}
	}
	return nil
}

func (tk TupleKey) ActionOrRelation() string {
	if tk.Action.IsAAction() {
		return tk.Action.String()
	} else if tk.Relation.IsARelation() {
		return tk.Relation.String()
	}
	return ""
}

func (tk TupleKey) WithUser(user string) TupleKey {
	return TupleKey{
		Subject:  NewEntityKey(UserType, user),
		Object:   tk.Object,
		Relation: tk.Relation,
		Action:   tk.Action,
	}
}

func (tk TupleKey) WithSubject(userType ObjectType, userName string) TupleKey {
	return TupleKey{
		Subject:  NewEntityKey(userType, userName),
		Object:   tk.Object,
		Relation: tk.Relation,
		Action:   tk.Action,
	}
}

func (tk TupleKey) WithSubjectID(userType ObjectType, userId int) TupleKey {
	return tk.WithSubject(userType, strconv.Itoa(userId))
}

func (tk TupleKey) WithObject(objectType ObjectType, objectName string) TupleKey {
	return TupleKey{
		Subject:  tk.Subject,
		Object:   NewEntityKey(objectType, objectName),
		Relation: tk.Relation,
		Action:   tk.Action,
	}
}

func (tk TupleKey) WithObjectID(objectType ObjectType, objectId int) TupleKey {
	return tk.WithObject(objectType, strconv.Itoa(objectId))
}

func (tk TupleKey) WithRelation(relation Relation) TupleKey {
	return TupleKey{
		Subject:  tk.Subject,
		Object:   tk.Object,
		Relation: relation,
		Action:   tk.Action,
	}
}

func (tk TupleKey) WithAction(action Action) TupleKey {
	return TupleKey{
		Subject:  tk.Subject,
		Object:   tk.Object,
		Relation: tk.Relation,
		Action:   action,
	}
}

func fromFGATupleKey(fgatk openfga.TupleKey) TupleKey {
	rel, _ := RelationString(*fgatk.Relation)
	act, _ := ActionString(*fgatk.Relation)
	return TupleKey{
		Subject:  NewEntityKeySplit(*fgatk.User),
		Object:   NewEntityKeySplit(*fgatk.Object),
		Relation: rel,
		Action:   act,
	}
}

func (tk TupleKey) FGATupleKey() openfga.TupleKey {
	fgatk := openfga.TupleKey{}
	if tk.Subject.Name != "" {
		fgatk.User = openfga.PtrString(tk.Subject.String())
	}
	if tk.Object.Name != "" {
		fgatk.Object = openfga.PtrString(tk.Object.String())
	}
	if tk.Action.IsAAction() {
		fgatk.Relation = openfga.PtrString(tk.Action.String())
	} else if tk.Relation.IsARelation() {
		fgatk.Relation = openfga.PtrString(tk.Relation.String())
	}
	return fgatk
}

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
	jj, err := json.Marshal(tkeys)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jj))
	return tkeys, nil
}

func rowGetString(row tlcsv.Row, col string) string {
	a, _ := row.Get(col)
	return a
}
