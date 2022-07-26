package handler

import "github.com/smf8/arvan-voucher/internal/app/model"

type RedemptionReportResponse struct {
	Length      int                `json:"length"`
	Redemptions []model.Redemption `json:"redemptions"`
}
