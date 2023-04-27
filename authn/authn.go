package authn

import (
	"context"

	"github.com/interline-io/transitland-server/auth"
)

type AuthnProvider interface {
	Check(context.Context, TupleKey) (bool, error)
	ListObjects(context.Context, TupleKey) ([]string, error)
}

type Checker struct {
	provider AuthnProvider
}

func NewChecker(p AuthnProvider) *Checker {
	return &Checker{
		provider: p,
	}
}

func (c *Checker) Check(ctx context.Context, tk TupleKey) (bool, error) {
	return c.provider.Check(ctx, tk)
}

func (c *Checker) Feeds(ctx context.Context, user auth.User) ([]int, error) {
	return []int{1, 2, 3}, nil
	// return c.provider.ListObjects(ctx, TupleKey{User: userKey, Object: "feed", Relation: "can_view"})
}
