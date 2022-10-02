package request

import "github.com/gookit/validate"

// CreateTransactionRequest 创建交易请求
type CreateTransactionRequest struct {
	OrderId     string  `json:"order_id" validate:"required|maxLen:32"`
	Amount      float64 `json:"amount" validate:"required|isFloat|gt:0.01"`
	NotifyUrl   string  `json:"notify_url" validate:"required"`
	Signature   string  `json:"signature"  validate:"required"`
	RedirectUrl string  `json:"redirect_url"`
	Timestamp   int64   `json:"timestamp" validate:"required"`
	Currency    string  `json:"currency"`
	ChainType   string  `json:"chain_type"`
}

func (r CreateTransactionRequest) Translates() map[string]string {
	return validate.MS{
		"OrderId":   "订单号",
		"Amount":    "支付金额",
		"NotifyUrl": "异步回调网址",
		"Signature": "签名",
		"Timestamp": "时间戳",
		"Currency":  "币种",
		"ChainType": "链类型",
	}
}

// OrderProcessingRequest 订单处理
type OrderProcessingRequest struct {
	Token              string
	Amount             float64
	TradeId            string
	BlockTransactionId string
}

type QueryTransactionRequest struct {
	TradeId   string `json:"trade_id" validate:"required|maxLen:32"`
	Timestamp int64  `json:"timestamp" validate:"required"`
	Signature string `json:"signature"  validate:"required"`
}

func (r QueryTransactionRequest) Translates() map[string]string {
	return validate.MS{
		"TradeId":   "交易号",
		"Timestamp": "时间戳",
	}
}
