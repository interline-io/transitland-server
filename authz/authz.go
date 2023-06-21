package authz

import (
	"context"
	"errors"
)

var ErrUnauthorized = errors.New("unauthorized")

type AuthnProvider interface {
	Users(context.Context, string) ([]*User, error)
	UserByID(context.Context, string) (*User, error)
}

type AuthzProvider interface {
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
