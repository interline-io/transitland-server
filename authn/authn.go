package authn

import "context"

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

func (c *Checker) Feeds(ctx context.Context, userKey string) ([]string, error) {
	return []string{"CT", "BA"}, nil
	// return c.provider.ListObjects(ctx, TupleKey{User: userKey, Object: "feed", Relation: "can_view"})
}
