package authz

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	openfga "github.com/openfga/go-sdk"
)

// For convenience
var FeedType = ObjectType_feed
var UserType = ObjectType_user
var TenantType = ObjectType_tenant
var GroupType = ObjectType_org
var FeedVersionType = ObjectType_feed_version

var ViewerRelation = Relation_viewer
var MemberRelation = Relation_member
var AdminRelation = Relation_admin
var ManagerRelation = Relation_manager
var ParentRelation = Relation_parent
var EditorRelation = Relation_editor

var CanEdit = Action_can_edit
var CanView = Action_can_view
var CanCreateFeedVersion = Action_can_create_feed_version
var CanDeleteFeedVersion = Action_can_delete_feed_version
var CanCreateFeed = Action_can_create_feed
var CanDeleteFeed = Action_can_delete_feed
var CanSetGroup = Action_can_set_group
var CanCreateOrg = Action_can_create_org
var CanEditMembers = Action_can_edit_members
var CanDeleteOrg = Action_can_delete_org

func CheckAction(actions []Action, check Action) bool {
	for _, a := range actions {
		if a == check {
			return true
		}
	}
	return false
}

func RelationString(v string) (Relation, error) {
	return Relation(0), nil
}

func ActionString(v string) (Action, error) {
	return Action(0), nil
}

func ObjectTypeString(v string) (ObjectType, error) {
	return ObjectType(0), nil
}

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

func NewEntityID(t ObjectType, id int64) EntityKey {
	return EntityKey{Type: t, Name: strconv.Itoa(int(id))}
}

func (ek EntityKey) ID() int64 {
	v, _ := strconv.Atoi(ek.Name)
	return int64(v)
}

func (ek EntityKey) String() string {
	if ek.Name == "" {
		return ek.Type.String()
	}
	return fmt.Sprintf("%s:%s", ek.Type.String(), ek.Name)
}

func IsRelation(Relation) bool {
	return true
}

func IsAction(Action) bool {
	return true
}

func IsObjectType(ObjectType) bool {
	return true
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
	if IsRelation(tk.Relation) {
		r = "|relation:" + tk.Relation.String()
	} else if IsAction(tk.Action) {
		r = "|action:" + tk.Action.String()
	}
	return fmt.Sprintf("%s|%s%s", tk.Subject.String(), tk.Object.String(), r)
}

func (tk TupleKey) IsValid() bool {
	return tk.Validate() == nil
}

func (tk TupleKey) Validate() error {
	if tk.Subject.Name != "" && !IsObjectType(tk.Subject.Type) {
		return errors.New("invalid user type")
	}
	if tk.Object.Name != "" && !IsObjectType(tk.Object.Type) {
		return errors.New("invalid object type")
	}
	if tk.Subject.Name == "" && tk.Object.Name == "" {
		return errors.New("user name or object name is required")
	}
	if tk.Subject.Name != "" && tk.Object.Name != "" {
		if tk.Action == 0 && !IsRelation(tk.Relation) {
			return errors.New("invalid relation")
		}
		if tk.Relation == 0 && !IsAction(tk.Action) {
			return errors.New("invalid action")
		}
	}
	return nil
}

func (tk TupleKey) ActionOrRelation() string {
	if IsAction(tk.Action) {
		return tk.Action.String()
	} else if IsRelation(tk.Relation) {
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

func (tk TupleKey) WithSubjectID(userType ObjectType, userId int64) TupleKey {
	return tk.WithSubject(userType, strconv.Itoa(int(userId)))
}

func (tk TupleKey) WithObject(objectType ObjectType, objectName string) TupleKey {
	return TupleKey{
		Subject:  tk.Subject,
		Object:   NewEntityKey(objectType, objectName),
		Relation: tk.Relation,
		Action:   tk.Action,
	}
}

func (tk TupleKey) WithObjectID(objectType ObjectType, objectId int64) TupleKey {
	return tk.WithObject(objectType, strconv.Itoa(int(objectId)))
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
	if IsAction(tk.Action) {
		fgatk.Relation = openfga.PtrString(tk.Action.String())
	} else if IsRelation(tk.Relation) {
		fgatk.Relation = openfga.PtrString(tk.Relation.String())
	}
	return fgatk
}
