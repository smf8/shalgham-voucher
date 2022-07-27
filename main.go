package main

import (
	"github.com/sirupsen/logrus"
	"github.com/smf8/arvan-voucher/internal/app/config"
	"github.com/smf8/arvan-voucher/internal/app/handler"
	"github.com/smf8/arvan-voucher/internal/app/model"
	"github.com/smf8/arvan-voucher/internal/app/wallet"
	"github.com/smf8/arvan-voucher/pkg/database"
	"github.com/smf8/arvan-voucher/pkg/log"
	"github.com/smf8/arvan-voucher/pkg/redis"
	"github.com/smf8/arvan-voucher/pkg/router"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.New()

	log.SetupLogger(cfg.LogLevel)

	app := router.New(cfg.Server)

	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logrus.Fatalf("database failed: %s", err.Error())
	}

	rc, redisCloser := redis.New(cfg.Redis)

	defer redisCloser()

	voucherRepo := &model.SQLVoucherRepo{DB: db}

	voucherCache := model.NewInMemoryVoucherCache(voucherRepo)

	if err := voucherCache.Start(cfg.VoucherCache.CronPattern); err != nil {
		logrus.Fatalf("voucher cache failed: %s", err)
	}

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
	app.Get("/healthz", handler.CheckHealth)
	api := app.Group("/api")
	api.Post("/vouchers", voucherHandler.Save)
	api.Get("/vouchers", voucherHandler.GetVoucher)
	api.Post("/vouchers/redeem", redeemerHandler.RedeemVoucher)
	api.Get("/vouchers/report", voucherHandler.Report)

	go func() {
		if err := app.Listen(cfg.Server.Port); err != nil {
			logrus.Fatalf("http server failed: %s", err.Error())
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	s := <-sig
	logrus.Infof("signal %s received\n", s)

	if err = app.Shutdown(); err != nil {
		logrus.Errorf("failed to shutdown server: %s", err.Error())
	}
}
