package authz

import (
	"errors"
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

type TupleKey struct {
	UserType   ObjectType `json:"user_type"`
	UserName   string     `json:"user_name"`
	ObjectType ObjectType `json:"object_type"`
	ObjectName string     `json:"object_name"`
	Action     Action     `json:"action"`
	Relation   Relation   `json:"relation"`
}

func (tk TupleKey) String() string {
	r := "relation:" + tk.Relation.String()
	if tk.Action.IsAAction() {
		r = "action:" + tk.Action.String()
	}
	return fmt.Sprintf("%s:%s|%s:%s|%s", tk.UserType.String(), tk.UserName, tk.ObjectType.String(), tk.ObjectName, r)
}

func (tk TupleKey) UserID() int {
	v, _ := strconv.Atoi(tk.UserName)
	return v
}

func (tk TupleKey) ObjectID() int {
	return tk.ObjectID()
}

func (tk TupleKey) IsValid() bool {
	return tk.Validate() == nil
}

func (tk TupleKey) Validate() error {
	if tk.UserName != "" && !tk.UserType.IsAObjectType() {
		return errors.New("invalid user type")
	}
	if tk.ObjectName != "" && !tk.ObjectType.IsAObjectType() {
		return errors.New("invalid object type")
	}
	if tk.UserName == "" && tk.ObjectName == "" {
		return errors.New("user name or object name is required")
	}
	if tk.UserName != "" && tk.ObjectName != "" {
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

func (tk TupleKey) WithUserName(user string) TupleKey {
	return TupleKey{
		UserType:   UserType,
		UserName:   user,
		ObjectType: tk.ObjectType,
		ObjectName: tk.ObjectName,
		Relation:   tk.Relation,
		Action:     tk.Action,
	}
}

func (tk TupleKey) WithUser(userType ObjectType, userName string) TupleKey {
	return TupleKey{
		UserType:   userType,
		UserName:   userName,
		ObjectType: tk.ObjectType,
		ObjectName: tk.ObjectName,
		Relation:   tk.Relation,
		Action:     tk.Action,
	}
}

func (tk TupleKey) WithUserID(userType ObjectType, userId int) TupleKey {
	return tk.WithUser(userType, strconv.Itoa(userId))
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

func (tk TupleKey) WithObjectID(objectType ObjectType, objectId int) TupleKey {
	return tk.WithObject(objectType, strconv.Itoa(objectId))
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

func tkObjectIds(tks []TupleKey) []int {
	var ret []int
	for _, tk := range tks {
		ret = append(ret, tk.ObjectID())
	}
	return ret
}
