package model

import (
	"context"
	"sync"
	"testing"

	"github.com/smf8/arvan-voucher/internal/app/config"
	"github.com/smf8/arvan-voucher/pkg/database"
	"github.com/smf8/arvan-voucher/pkg/redis"
	"github.com/stretchr/testify/suite"
)

type VoucherTestSuite struct {
	suite.Suite
	voucherRepo          *SQLVoucherRepo
	voucherRemainderRepo *RedisVoucherRemainderRepo
	redemptionRepo       *SQLRedemptionRepo
}

func (suite *VoucherTestSuite) SetupSuite() {
	cfg := config.New()

	db, err := database.NewConnection(cfg.Database)

	suite.NoError(err)

	rc, _ := redis.New(cfg.Redis)

	suite.voucherRemainderRepo = &RedisVoucherRemainderRepo{Redis: rc}
	suite.voucherRepo = &SQLVoucherRepo{DB: db}
	suite.redemptionRepo = &SQLRedemptionRepo{DB: db}
}

func (suite *VoucherTestSuite) SetupTest() {
	suite.NoError(suite.voucherRemainderRepo.Redis.FlushDB(context.Background()).Err())
	suite.voucherRepo.DB.Exec("truncate vouchers, redemptions")
}

func (suite *VoucherTestSuite) TestVoucherRemainder() {
	// scenario: 50 concurrent users try to call Use() 800 times.
	//having 1000 as initial limit, we must have 200 as redis value after the calls.
	suite.Run("concurrent run", func() {
		wg := &sync.WaitGroup{}

		concurrentUsers := 50
		limit := 1000
		calls := 800
		finalValue := limit - calls

		// setup redis
		suite.NoError(suite.voucherRemainderRepo.Redis.Set(context.Background(),
			suite.voucherRemainderRepo.voucherRemainderKey("test"),
			limit, 0).Err())

		callsLock := sync.Mutex{}

		callCounter := new(int)

		for i := 0; i < concurrentUsers; i++ {
			wg.Add(1)

			go func(i *int) {
				for {
					callsLock.Lock()

					if *i >= calls {
						callsLock.Unlock()
						break
					}

					*i++

					callsLock.Unlock()

					value, err := suite.voucherRemainderRepo.Use(context.Background(), "test")
					suite.NoError(err)
					suite.True(value)
				}

				wg.Done()
			}(callCounter)
		}

		wg.Wait()

		value, err := suite.voucherRemainderRepo.Get(context.Background(), "test")
		suite.NoError(err)

		suite.Equal(finalValue, value)
	})
}

func (suite *VoucherTestSuite) TestVoucherRepo() {
	suite.Run("Create/Read", func() {
		v1 := &Voucher{
			Code:   "test_v1",
			Amount: 1000,
			Limit:  10,
		}

		suite.NoError(suite.voucherRepo.Save(context.Background(), v1))

		result, err := suite.voucherRepo.Find(context.Background(), v1.Code)
		suite.NoError(err)

		suite.Equal(v1.Amount, result.Amount)
		suite.Equal(v1.Limit, result.Limit)
	})

	suite.Run("Delete", func() {
		v1 := &Voucher{
			Code:   "test_v2",
			Amount: 1001,
			Limit:  11,
		}

		suite.NoError(suite.voucherRepo.Save(context.Background(), v1))

		suite.NoError(suite.voucherRepo.Delete(context.Background(), v1.Code))

		result, err := suite.voucherRepo.Find(context.Background(), v1.Code)

		suite.Error(err)
		suite.Nil(result)
	})
}

func TestVoucherTestSuite(t *testing.T) {
	suite.Run(t, new(VoucherTestSuite))
}

func BenchmarkRedisVoucherRemainderRepo_Use(b *testing.B) {
	cfg := config.New()
	rc, _ := redis.New(cfg.Redis)

	repo := &RedisVoucherRemainderRepo{Redis: rc}

	if err := repo.Redis.Set(context.Background(),
		repo.voucherRemainderKey("bench"), 10000, 0).
		Err(); err != nil {
		b.Errorf("setup failed: %s", err.Error())
	}

	b.ResetTimer()

	//b.SetParallelism(1000)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := repo.Use(context.Background(), "bench")
			if err != nil {
				b.Errorf("use failed: %s", err.Error())
			}
		}
	})
}
