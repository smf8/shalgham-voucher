package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/smf8/arvan-voucher/internal/app/model"
	"github.com/smf8/arvan-voucher/internal/app/wallet"
	"net/http"
)

type Redeemer struct {
	VoucherRemainderRepo model.VoucherRemainderRepo
	RedemptionRepo       model.RedemptionRepo
	Client               *wallet.Client
	VoucherCache         *model.InMemoryVoucherCache
}

func (r *Redeemer) RedeemVoucher(c *fiber.Ctx) error {
	request := &RedeemRequest{}

	if err := c.BodyParser(request); err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	if err := request.validate(); err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	// order of operations:
	// 1- check if voucher is in local cache (to avoid any unnecessary network calls)
	// 2- check redis for voucher remainder status
	// 3- write voucher transaction into SQL
	// 4- apply voucher transaction to wallet through API

	voucherAmount, err := r.VoucherCache.GetVoucherAmount(request.Code)
	if err != nil {
		logrus.Debugf("voucher cache get failed: %s", err.Error())

		return c.SendStatus(http.StatusNotFound)
	}

	valid, err := r.VoucherRemainderRepo.Use(c.UserContext(), request.Code)
	if err != nil {
		logrus.Errorf("voucher remainder repo failed: %s", err.Error())

		return c.SendStatus(http.StatusInternalServerError)
	}

	if !valid {
		logrus.Debugf("voucher remainder repo: not valid")

		// for sake of report accuracy we should set transactionErr to revert Decr
		// but since number of requests for invalid vouchers are very high,
		//we can skip calling revert Incr in favor of Redis performance
		return c.SendStatus(http.StatusNotFound)
	}

	// revert is an inline function to revert parts of transaction before the failure
	revert := func(code string, redemption *model.Redemption) {
		if err := r.VoucherRemainderRepo.Revert(c.UserContext(), request.Code); err != nil {
			logrus.Errorf("redeem remainder revert failed: %s", err.Error())
		}

		if redemption != nil {
			if err := r.RedemptionRepo.Delete(c.UserContext(), redemption); err != nil {
				logrus.Errorf("redeem redemption revert failed: %s", err.Error())
			}
		}
	}

	redemption := &model.Redemption{
		VoucherCode: request.Code,
		Redeemer:    request.PhoneNumber,
	}

	err = r.RedemptionRepo.Create(c.UserContext(), redemption)
	if err != nil {
		logrus.Errorf("redemption create failed: %s", err.Error())

		revert(request.Code, nil)

		return c.SendStatus(http.StatusInternalServerError)
	}

	err = r.Client.ApplyTransaction(request.PhoneNumber, voucherAmount)
	if err != nil {
		logrus.Errorf("transaction apply failed: %s", err.Error())

		revert(request.Code, redemption)

		return c.SendStatus(http.StatusInternalServerError)
	}

	return c.SendStatus(http.StatusOK)
}
