package admin

import (
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type billingModeResponse struct {
	Currency     string  `json:"currency"`
	CurrencyName string  `json:"currency_name"`
	USDToCNYRate float64 `json:"usd_to_cny_rate"`
	CanConvert   bool    `json:"can_convert"`
}

type updateBillingModeRequest struct {
	Currency        string  `json:"currency" binding:"required"`
	USDToCNYRate    float64 `json:"usd_to_cny_rate"`
	ConvertExisting bool    `json:"convert_existing"`
	Confirmation    string  `json:"confirmation"`
}

func (h *SettingHandler) SetBillingModeService(billingModeService *service.BillingModeService) {
	h.billingModeService = billingModeService
}

func (h *SettingHandler) GetBillingMode(c *gin.Context) {
	if h.billingModeService == nil {
		response.InternalError(c, "billing mode service is not configured")
		return
	}
	billing := h.billingModeService.Current()
	rate := billing.USDToCNYRate
	if rate <= 0 {
		rate = 7.2
	}
	response.Success(c, billingModeResponse{
		Currency:     billing.CurrencyCode(),
		CurrencyName: map[string]string{"USD": "美元", "CNY": "人民币"}[billing.CurrencyCode()],
		USDToCNYRate: rate,
		CanConvert:   true,
	})
}

func (h *SettingHandler) UpdateBillingMode(c *gin.Context) {
	if h.billingModeService == nil {
		response.InternalError(c, "billing mode service is not configured")
		return
	}
	var req updateBillingModeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid billing mode request")
		return
	}
	result, err := h.billingModeService.Update(c.Request.Context(), strings.TrimSpace(req.Currency), req.USDToCNYRate, req.ConvertExisting, req.Confirmation)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	billing := h.billingModeService.Current()
	rate := billing.USDToCNYRate
	if rate <= 0 {
		rate = 7.2
	}
	response.Success(c, gin.H{
		"currency":        billing.CurrencyCode(),
		"currency_name":   map[string]string{"USD": "美元", "CNY": "人民币"}[billing.CurrencyCode()],
		"usd_to_cny_rate": rate,
		"rows_converted":  result.RowsConverted,
		"tables":          result.Tables,
	})
}
