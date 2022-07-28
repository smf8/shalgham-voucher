package model

import (
	"context"
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"sync"
)

type InMemoryVoucherCache struct {
	lock  sync.RWMutex
	cache *cache.Cache
	repo  VoucherRepo
}

func NewInMemoryVoucherCache(repo VoucherRepo) *InMemoryVoucherCache {
	c := cache.New(cache.NoExpiration, cache.NoExpiration)

	return &InMemoryVoucherCache{
		cache: c,
		repo:  repo,
	}
}

func (c *InMemoryVoucherCache) GetVoucherAmount(voucherCode string) (float64, error) {
	c.lock.RLock()

	value, found := c.cache.Get(voucherCode)
	if !found {
		c.lock.RUnlock()

		return 0, fmt.Errorf("voucher cache: voucher with code %s not found", voucherCode)
	}

	c.lock.RUnlock()

	result, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("voucher cache: invalid type for voucher: %t", value)
	}

	return result, nil

}

func (c *InMemoryVoucherCache) Start(cronPattern string) error {
	if err := c.fetch(); err != nil {
		return err
	}

	cronJob := cron.New()
	if _, err := cronJob.AddFunc(cronPattern, func() {
		if err := c.fetch(); err != nil {
			logrus.Errorf("voucher cache: cron job fetch failed: %s", err.Error())
		}
	}); err != nil {
		return err
	}

	cronJob.Start()

	return nil
}

func (c *InMemoryVoucherCache) fetch() error {
	vouchers, err := c.repo.FindAll(context.Background())
	if err != nil {
		return fmt.Errorf("voucher cache: failed to find all vouchers: %s", err.Error())
	}

	// since the interval after flushing and inserting new items
	// cause inconsistent data state. we use a lock to safely flush and re-insert data.
	c.lock.Lock()
	c.cache.Flush()

	for _, v := range vouchers {
		c.cache.Set(v.Code, v.Amount, cache.NoExpiration)
	}

	c.lock.Unlock()

	return nil
}
