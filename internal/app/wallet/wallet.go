package wallet

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/smf8/arvan-voucher/internal/app/config"
	"github.com/smf8/arvan-voucher/pkg/router"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	timeout    time.Duration
	debug      bool
	baseURL    string
	httpClient *fiber.Client
}

type TransactionRequest struct {
	PhoneNumber string  `json:"phone_number"`
	Amount      float64 `json:"amount"`
}

func NewClient(cfg config.WalletClient) *Client {
	httpClient := fiber.AcquireClient()

	return &Client{
		timeout:    cfg.Timeout,
		debug:      cfg.Debug,
		baseURL:    cfg.BaseURL,
		httpClient: httpClient,
	}
}

func (c *Client) ApplyTransaction(phoneNumber string, transactionValue float64) error {
	agent := c.httpClient.Post(c.baseURL + "/api/transactions").Timeout(c.timeout)

	if c.debug {
		agent.Debug()
	}

	req := &TransactionRequest{
		PhoneNumber: phoneNumber,
		Amount:      transactionValue,
	}

	agent.JSON(req)

	agent.Set(router.RequestTimeoutHeaderKey, c.timeout.String())

	responseCode, responseBody, errs := agent.String()
	if len(errs) != 0 {
		errorMessages := make([]string, len(errs))
		for i, err := range errs {
			errorMessages[i] = err.Error()
		}

		return fmt.Errorf("http request failed: %s", strings.Join(errorMessages, ", "))
	}

	if responseCode != http.StatusOK {
		return fmt.Errorf("http request non 200 code: %s", responseBody)
	}

	return nil
}
