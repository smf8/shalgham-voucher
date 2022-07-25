package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"net/http"
	"voucher/model"
	"voucher/wallet"
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
		return c.SendStatus(http.StatusNotFound)
	}

	valid, err := r.VoucherRemainderRepo.Use(c.UserContext(), request.Code)
	if err != nil {
		logrus.Errorf("voucher remainder repo failed: %s", err.Error())

		return c.SendStatus(http.StatusInternalServerError)
	}

	if !valid {
		return c.SendStatus(http.StatusNotFound)
	}

	var transactionErr error

	// if transaction fails. we revert counter decrease
	defer func() {
		if transactionErr != nil {
			r.VoucherRemainderRepo.Revert(c.UserContext(), request.Code)
		}
	}()

	redemption := &model.Redemption{
		VoucherCode: request.Code,
		Redeemer:    request.PhoneNumber,
	}

	transactionErr = r.RedemptionRepo.Create(redemption)
	if transactionErr != nil {
		logrus.Errorf("redemption create failed: %s", transactionErr.Error())

		return c.SendStatus(http.StatusInternalServerError)
	}

	transactionErr = r.Client.ApplyTransaction(request.PhoneNumber, voucherAmount)
	if err != nil {
		logrus.Errorf("transaction apply failed: %s", transactionErr.Error())

		return c.SendStatus(http.StatusInternalServerError)
	}

	return c.SendStatus(http.StatusOK)
}
