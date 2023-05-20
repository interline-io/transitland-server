package authz

import (
	"fmt"
	"strconv"
	"strings"

	openfga "github.com/openfga/go-sdk"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type EntityKey struct {
	Type ObjectType `json:"type"`
	Name string     `json:"name"`
}

func NewEntityKey(t ObjectType, name string) EntityKey {
	return EntityKey{Type: t, Name: name}
}

func NewEntityID(t ObjectType, id int) EntityKey {
	return EntityKey{Type: t, Name: strconv.Itoa(id)}
}

func (ek EntityKey) ID() int {
	v, _ := strconv.Atoi(ek.Name)
	return v
}

func (ek EntityKey) String() string {
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
	r := "relation:" + tk.Relation.String()
	if tk.Action.IsAAction() {
		r = "action:" + tk.Action.String()
	}
	return fmt.Sprintf("%s|%s|%s", tk.Subject.String(), tk.Object.String(), r)
}

func (tk TupleKey) IsValid() bool {
	return tk.Validate() == nil
}

func (tk TupleKey) Validate() error {
	// if tk.SubjectName != "" && !tk.SubjectType.IsAObjectType() {
	// 	return errors.New("invalid user type")
	// }
	// if tk.ObjectName != "" && !tk.ObjectType.IsAObjectType() {
	// 	return errors.New("invalid object type")
	// }
	// if tk.SubjectName == "" && tk.ObjectName == "" {
	// 	return errors.New("user name or object name is required")
	// }
	// if tk.SubjectName != "" && tk.ObjectName != "" {
	// 	if tk.Action == 0 && !tk.Relation.IsARelation() {
	// 		return errors.New("invalid relation")
	// 	}
	// 	if tk.Relation == 0 && !tk.Action.IsAAction() {
	// 		return errors.New("invalid action")
	// 	}
	// }
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
	okeys := csplit(*fgatk.Object)
	ukeys := csplit(*fgatk.User)
	rel, _ := RelationString(*fgatk.Relation)
	act, _ := ActionString(*fgatk.Relation)
	return TupleKey{
		Subject:  NewEntityKey(ukeys.Type, ukeys.Name),
		Object:   NewEntityKey(okeys.Type, okeys.Name),
		Relation: rel,
		Action:   act,
	}
}

func (tk TupleKey) FGATupleKey() openfga.TupleKey {
	fgatk := openfga.TupleKey{}
	if tk.Subject.Name != "" {
		fgatk.User = openfga.PtrString(cunsplit(tk.Subject.Type, tk.Subject.Name))
	}
	if tk.Object.Name != "" {
		fgatk.Object = openfga.PtrString(cunsplit(tk.Object.Type, tk.Object.Name))
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
