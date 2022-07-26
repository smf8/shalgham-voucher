package main

import (
	"github.com/sirupsen/logrus"
	"github.com/smf8/arvan-voucher/internal/app/config"
	"github.com/smf8/arvan-voucher/internal/app/handler"
	"github.com/smf8/arvan-voucher/internal/app/model"
	"github.com/smf8/arvan-voucher/internal/app/wallet"
	"github.com/smf8/arvan-voucher/pkg/database"
	"github.com/smf8/arvan-voucher/pkg/redis"
	"github.com/smf8/arvan-voucher/pkg/router"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.New()

	app := router.New(cfg.Server)

	api := app.Group("/api")

	go func() {
		if err := app.Listen(cfg.Server.Port); err != nil {
			logrus.Fatalf("http server failed: %s", err.Error())
		}
	}()

	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logrus.Fatalf("database failed: %s", err.Error())
	}

	rc, redisCloser := redis.New(cfg.Redis)

	defer redisCloser()

	voucherRepo := &model.SQLVoucherRepo{DB: db}
	voucherCache := model.NewInMemoryVoucherCache(voucherRepo)

	redemptionRepo := &model.SQLRedemptionRepo{DB: db}
	voucherRemainderRepo := &model.RedisVoucherRemainderRepo{Redis: rc}

	voucherHandler := handler.Voucher{
		VoucherRepo:          voucherRepo,
		RedemptionRepo:       redemptionRepo,
		VoucherRemainderRepo: voucherRemainderRepo,
	}

	redeemerHandler := handler.Redeemer{
		VoucherRemainderRepo: voucherRemainderRepo,
		RedemptionRepo:       redemptionRepo,
		Client:               wallet.NewClient(cfg.WalletClient),
		VoucherCache:         voucherCache,
	}

	// Register Routes
	api.Get("/voucher", voucherHandler.GetVoucher)
	api.Post("/voucher", voucherHandler.Save)
	api.Post("/voucher/redeem", redeemerHandler.RedeemVoucher)
	api.Get("/voucher/report", voucherHandler.Report)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	s := <-sig
	logrus.Infof("signal %s received\n", s)

	if err = app.Shutdown(); err != nil {
		logrus.Errorf("failed to shutdown server: %s", err.Error())
	}
}
