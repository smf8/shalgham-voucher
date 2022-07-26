package handler

import (
	"errors"
	"regexp"
)

var mobileRegex = regexp.MustCompile(`^(\+98)[0-9]{10}$`)

type RedeemRequest struct {
	PhoneNumber string `json:"phone_number"`
	Code        string `json:"code"`
}

type VoucherReportRequest struct {
	Code   string `json:"code"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

func (r *RedeemRequest) validate() error {
	if !mobileRegex.MatchString(r.PhoneNumber) {
		return errors.New("invalid phone number format. use +98xxxxxxxxxx")
	}

	return nil
}
