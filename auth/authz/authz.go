package authz

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/interline-io/transitland-server/internal/generated/azpb"
	openfga "github.com/openfga/go-sdk"
)

var ErrUnauthorized = errors.New("unauthorized")

type UserProvider interface {
	Users(context.Context, string) ([]*azpb.User, error)
	UserByID(context.Context, string) (*azpb.User, error)
}

type FGAProvider interface {
	Check(context.Context, TupleKey, ...TupleKey) (bool, error)
	ListObjects(context.Context, TupleKey) ([]TupleKey, error)
	GetObjectTuples(context.Context, TupleKey) ([]TupleKey, error)
	WriteTuple(context.Context, TupleKey) error
	ReplaceTuple(context.Context, TupleKey) error
	DeleteTuple(context.Context, TupleKey) error
}

type AuthzConfig struct {
	Auth0Domain       string
	Auth0ClientID     string
	Auth0ClientSecret string
	FGAStoreID        string
	FGAModelID        string
	FGAEndpoint       string
	FGALoadModelFile  string
	FGALoadTupleFile  string
	GlobalAdmin       string
}

// For less typing

type Action = azpb.Action
type ObjectType = azpb.ObjectType
type Relation = azpb.Relation

var FeedType = azpb.FeedType
var UserType = azpb.UserType
var TenantType = azpb.TenantType
var GroupType = azpb.GroupType
var FeedVersionType = azpb.FeedVersionType

var ViewerRelation = azpb.ViewerRelation
var MemberRelation = azpb.MemberRelation
var AdminRelation = azpb.AdminRelation
var ManagerRelation = azpb.ManagerRelation
var ParentRelation = azpb.ParentRelation
var EditorRelation = azpb.EditorRelation

var CanEdit = azpb.CanEdit
var CanView = azpb.CanView
var CanCreateFeedVersion = azpb.CanCreateFeedVersion
var CanDeleteFeedVersion = azpb.CanDeleteFeedVersion
var CanCreateFeed = azpb.CanCreateFeed
var CanDeleteFeed = azpb.CanDeleteFeed
var CanSetGroup = azpb.CanSetGroup
var CanCreateOrg = azpb.CanCreateOrg
var CanEditMembers = azpb.CanEditMembers
var CanDeleteOrg = azpb.CanDeleteOrg

type EntityKey = azpb.EntityKey
type TupleKey = azpb.TupleKey

func NewEntityKey(t ObjectType, name string) EntityKey {
	return azpb.NewEntityKey(t, name)
}

func NewEntityKeySplit(v string) EntityKey {
	ret := EntityKey{}
	a := strings.Split(v, ":")
	if len(a) > 1 {
		ret.Type, _ = azpb.ObjectTypeString(a[0])
		ret.Name = a[1]
	} else if len(a) > 0 {
		ret.Type, _ = azpb.ObjectTypeString(a[0])
	}
	return ret
}

func NewEntityID(t ObjectType, id int64) EntityKey {
	return azpb.NewEntityKey(t, strconv.Itoa(int(id)))
}

func NewTupleKey() TupleKey { return TupleKey{} }

func FromFGATupleKey(fgatk openfga.TupleKey) TupleKey {
	rel, _ := azpb.RelationString(*fgatk.Relation)
	act, _ := azpb.ActionString(*fgatk.Relation)
	return TupleKey{
		Subject:  NewEntityKeySplit(*fgatk.User),
		Object:   NewEntityKeySplit(*fgatk.Object),
		Relation: rel,
		Action:   act,
	}
}

func ToFGATupleKey(tk TupleKey) openfga.TupleKey {
	fgatk := openfga.TupleKey{}
	if tk.Subject.Name != "" {
		fgatk.User = openfga.PtrString(tk.Subject.String())
	}
	if tk.Object.Name != "" {
		fgatk.Object = openfga.PtrString(tk.Object.String())
	}
	if azpb.IsAction(tk.Action) {
		fgatk.Relation = openfga.PtrString(tk.Action.String())
	} else if azpb.IsRelation(tk.Relation) {
		fgatk.Relation = openfga.PtrString(tk.Relation.String())
	}
	return fgatk
}
