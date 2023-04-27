package authn

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/internal/ecache"
)

type AuthnProvider interface {
	Check(context.Context, TupleKey) (bool, error)
	ListObjects(context.Context, TupleKey) ([]string, error)
}

type Checker struct {
	provider  AuthnProvider
	feedCache *ecache.Cache[int]
	fvidCache *ecache.Cache[int]
}

func NewChecker(p AuthnProvider, redisClient *redis.Client) *Checker {
	return &Checker{
		provider:  p,
		feedCache: ecache.NewCache[int](redisClient, "checker:feeds"),
		fvidCache: ecache.NewCache[int](redisClient, "checker:fvids"),
	}
}

func (c *Checker) Check(ctx context.Context, tk TupleKey) (bool, error) {
	return c.provider.Check(ctx, tk)
}

func (c *Checker) Feeds(ctx context.Context, user auth.User) ([]int, error) {
	// Check cache

	// Use ListObjects, map back to feed IDs using current_feeds.authn_id
	// return c.provider.ListObjects(ctx, TupleKey{User: userKey, Object: "feed", Relation: "can_view"})

	return []int{1, 2, 3}, nil
}
