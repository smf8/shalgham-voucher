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

func (r *RedeemRequest) validate() error {
	if !mobileRegex.MatchString(r.PhoneNumber) {
		return errors.New("invalid phone number format. use +98xxxxxxxxxx")
	}

	return nil
}
