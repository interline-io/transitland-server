package authz

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/interline-io/transitland-server/internal/generated/azpb"
	openfga "github.com/openfga/go-sdk"
)

// For convenience
type Action = azpb.Action
type ObjectType = azpb.ObjectType
type Relation = azpb.Relation

var FeedType = azpb.ObjectType_feed
var UserType = azpb.ObjectType_user
var TenantType = azpb.ObjectType_tenant
var GroupType = azpb.ObjectType_org
var FeedVersionType = azpb.ObjectType_feed_version

var ViewerRelation = azpb.Relation_viewer
var MemberRelation = azpb.Relation_member
var AdminRelation = azpb.Relation_admin
var ManagerRelation = azpb.Relation_manager
var ParentRelation = azpb.Relation_parent
var EditorRelation = azpb.Relation_editor

var CanEdit = azpb.Action_can_edit
var CanView = azpb.Action_can_view
var CanCreateFeedVersion = azpb.Action_can_create_feed_version
var CanDeleteFeedVersion = azpb.Action_can_delete_feed_version
var CanCreateFeed = azpb.Action_can_create_feed
var CanDeleteFeed = azpb.Action_can_delete_feed
var CanSetGroup = azpb.Action_can_set_group
var CanCreateOrg = azpb.Action_can_create_org
var CanEditMembers = azpb.Action_can_edit_members
var CanDeleteOrg = azpb.Action_can_delete_org

func RelationString(v string) (Relation, error) {
	if a, ok := azpb.Relation_value[v]; ok {
		return Relation(a), nil
	}
	return Relation(0), errors.New("invalid relation")
}

func ActionString(v string) (Action, error) {
	if a, ok := azpb.Action_value[v]; ok {
		return Action(a), nil
	}
	return Action(0), errors.New("invalid action")
}

func ObjectTypeString(v string) (ObjectType, error) {
	if a, ok := azpb.ObjectType_value[v]; ok {
		return ObjectType(a), nil
	}
	return ObjectType(0), errors.New("invalid object type")
}

func IsRelation(v Relation) bool {
	_, ok := azpb.Relation_name[int32(v)]
	return ok && v > 0
}

func IsAction(v Action) bool {
	_, ok := azpb.Action_name[int32(v)]
	return ok && v > 0
}

func IsObjectType(v ObjectType) bool {
	_, ok := azpb.ObjectType_name[int32(v)]
	return ok && v > 0
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
