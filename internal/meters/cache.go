package meters

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-server/internal/ecache"
)

func init() {
	var _ MeterProvider = &CacheMeterProvider{}
}

type meterValueCache struct {
	Value float64
}

type CacheMeterProvider struct {
	cache    *ecache.Cache[meterValueCache]
	provider MeterProvider
}

func NewCacheMeterProvider(redisClient *redis.Client, provider MeterProvider) *CacheMeterProvider {
	return &CacheMeterProvider{
		provider: provider,
		cache:    ecache.NewCache[meterValueCache](redisClient, "cachemeter"),
	}
}

func (c *CacheMeterProvider) NewMeter(u MeterUser) ApiMeter {
	return &CacheMeter{
		userId: u.ID(),
		meter:  c.provider.NewMeter(u),
		cm:     c,
	}
}

func (c *CacheMeterProvider) Close() error {
	return c.provider.Close()
}

func (c *CacheMeterProvider) Flush() error {
	return c.provider.Flush()
}

type CacheMeter struct {
	userId string
	meter  ApiMeter
	cm     *CacheMeterProvider
}

func (c *CacheMeter) Meter(meterName string, value float64, extraDimensions Dimensions) error {
	return c.meter.Meter(meterName, value, extraDimensions)
}

func (c *CacheMeter) GetValue(meterName string, d time.Duration, dims Dimensions) (float64, bool) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s", c.userId, meterName)
	a, ok := c.cm.cache.Get(ctx, key)
	if !ok {
		a.Value, ok = c.meter.GetValue(meterName, d, dims)
		c.cm.cache.SetTTL(ctx, key, a, 10*time.Minute, 10*time.Minute)
	}
	return a.Value, ok
}

func (c *CacheMeter) AddDimension(meterName string, key string, value string) {
	c.meter.AddDimension(meterName, key, value)
}
