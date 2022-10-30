package model

import (
	"testing"

	"github.com/smf8/shalgham-voucher/internal/app/config"
	"github.com/smf8/shalgham-voucher/pkg/database"
	"github.com/stretchr/testify/suite"
)

type VoucherCacheTestSuite struct {
	suite.Suite
	voucherRepo  *SQLVoucherRepo
	voucherCache *InMemoryVoucherCache
}

func (suite *VoucherCacheTestSuite) SetupSuite() {
	cfg := config.New()

	db, err := database.NewConnection(cfg.Database)

	suite.NoError(err)

	suite.voucherRepo = &SQLVoucherRepo{DB: db}
	suite.voucherCache = NewInMemoryVoucherCache(suite.voucherRepo)

	suite.voucherCache.Start("@every 1s")
}

func (suite *VoucherCacheTestSuite) TestCache() {

	for i := 0; i < 10000000; i++ {
		go func() {
			_, err := suite.voucherCache.GetVoucherAmount("voucher_1")

			suite.NoError(err)
		}()
	}
}

func TestVoucherCacheSuite(t *testing.T) {
	suite.Run(t, new(VoucherCacheTestSuite))
}
