package handler

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/smf8/arvan-voucher/internal/app/model"
	"net/http"
)

type Voucher struct {
	VoucherRepo          model.VoucherRepo
	RedemptionRepo       model.RedemptionRepo
	VoucherRemainderRepo model.VoucherRemainderRepo
}

func (v *Voucher) Report(c *fiber.Ctx) error {
	request := &VoucherReportRequest{}

	if err := c.BodyParser(request); err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	redemptions, err := v.RedemptionRepo.FindRedemptions(request.Code, request.Limit, request.Offset)
	if err != nil {
		logrus.Errorf("find redemptions failed: %s", err.Error())

		return c.SendStatus(http.StatusInternalServerError)
	}

	response := RedemptionReportResponse{
		Length:      len(redemptions),
		Redemptions: redemptions,
	}

	return c.JSON(response)
}

func (v *Voucher) GetVoucher(c *fiber.Ctx) error {
	voucherCode := c.Query("code")
	if voucherCode == "" {
		return c.SendStatus(http.StatusBadRequest)
	}

	voucher, err := v.VoucherRepo.Find(voucherCode)
	if err != nil {
		if errors.Is(err, model.ErrRecordNotFound) {
			return c.SendStatus(http.StatusNotFound)
		}

		logrus.Errorf("voucher find failed: %s", err.Error())

		return c.SendStatus(http.StatusInternalServerError)
	}

	remaining, err := v.VoucherRemainderRepo.Get(c.UserContext(), voucherCode)
	if err != nil {
		logrus.Errorf("voucher remainder get failed: %s", err.Error())

		return c.SendStatus(http.StatusInternalServerError)
	}

	result := map[string]interface{}{
		"id":        voucher.ID,
		"code":      voucherCode,
		"amount":    voucher.Amount,
		"limit":     voucher.Limit,
		"remaining": remaining,
	}

	return c.JSON(result)
}

func (v *Voucher) Save(c *fiber.Ctx) error {
	voucherRequest := &model.Voucher{}

	if err := c.BodyParser(voucherRequest); err != nil {
		return c.Status(http.StatusBadRequest).SendString(err.Error())
	}

	if err := v.VoucherRepo.Save(voucherRequest); err != nil {
		logrus.Errorf("voucher save failed: %s", err.Error())

		return c.SendStatus(http.StatusInternalServerError)
	}

	if err := v.VoucherRemainderRepo.Create(
		c.UserContext(), voucherRequest.Code, voucherRequest.Limit); err != nil {
		logrus.Errorf("voucher remainder save failed: %s", err.Error())

		return c.SendStatus(http.StatusInternalServerError)
	}

	return c.Status(http.StatusCreated).JSON(voucherRequest)
}
