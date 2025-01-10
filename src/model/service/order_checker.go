package service

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/assimon/luuu/config"
	"github.com/assimon/luuu/model/data"
	"github.com/assimon/luuu/model/request"
	"github.com/assimon/luuu/mq"
	"github.com/assimon/luuu/mq/handle"
	"github.com/assimon/luuu/telegram"
	"github.com/assimon/luuu/util/http_client"
	"github.com/assimon/luuu/util/json"
	"github.com/go-resty/resty/v2"
	"github.com/hibiken/asynq"
	"github.com/shopspring/decimal"
	"github.com/webx-top/com"
	"github.com/webx-top/echo/param"
	"gopkg.in/yaml.v3"
)

var (
	defs       []*OrderCheckerDef
	defsInited bool
	dmu        sync.RWMutex
	chkr       OrderChecker
	once       sync.Once
)

func init() {
	config.OnInitialize(func(c *config.Config) error {
		if len(c.CheckerDefPath) == 0 {
			return nil
		}
		return ParseConfig(c.CheckerDefPath)
	})
}

func ParseConfig(configFile string) error {
	var defs []*OrderCheckerDef
	b, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, &defs)
	if err != nil {
		return err
	}
	SetDefs(defs)
	return nil
}

func DefsInited() bool {
	dmu.Lock()
	r := defsInited
	dmu.Unlock()
	return r
}

func SetDefs(_defs []*OrderCheckerDef) {
	dmu.Lock()
	defs = _defs
	defsInited = true
	dmu.Unlock()
}

func Defs() []*OrderCheckerDef {
	dmu.RLock()
	r := defs
	dmu.RUnlock()
	return r
}

func initChecker() {
	chkr = NewDefaultCheck(Defs())
}

func Checker() OrderChecker {
	once.Do(initChecker)
	return chkr
}

type OrderChecker interface {
	Check(token string) error
}

func NewDefaultCheck(defs []*OrderCheckerDef) *defaultCheck {
	return &defaultCheck{
		defs:   defs,
		client: http_client.GetHttpClient(),
	}
}

func NewTronscanapiDef() *OrderCheckerDef {
	d := NewCheckerDef(UsdtTrc20ApiUri)
	d.QueryParams = map[string]string{
		"sort":            "-timestamp",
		"limit":           "50",
		"start":           "0",
		"direction":       "2",
		"db_version":      "1",
		"trc20Id":         "{trc20ContractAddress}",
		"address":         "{token}",
		"start_timestamp": "{startTime}",
		"end_timestamp":   "{endTime}",
	}
	d.ListKeyName = `data`
	d.CountKeyName = `page_size`
	d.ItemKeyName.ToToken = `to`
	d.ItemKeyName.Status = `contract_ret`
	d.ItemKeyName.Amount = `amount`
	d.ItemKeyName.Timestamp = `block_timestamp`
	d.ItemKeyName.TransactionId = `hash`
	d.AmountDivisor = 1000000
	d.ItemSuccessValue = `SUCCESS`
	return d
}

func NewCheckerDef(baseURL string) *OrderCheckerDef {
	return &OrderCheckerDef{
		BaseURL:          baseURL,
		QueryParams:      map[string]string{},
		Headers:          map[string]string{},
		ItemKeyName:      &ItemKeyName{},
		ItemSuccessValue: `SUCCESS`,
	}
}

type defaultCheck struct {
	defs   []*OrderCheckerDef
	client *resty.Client
}

func (d *defaultCheck) Check(token string) (err error) {
	startTime := time.Now().Add(-24 * time.Hour).UnixMilli()
	endTime := time.Now().UnixMilli()
	for _, def := range d.defs {
		err = d.check(def, token, startTime, endTime)
		if err == nil {
			return
		}
		if !errors.Is(err, ErrAPI) {
			return
		}
	}
	return
}

var (
	ErrAPI                = errors.New(`API Error`)
	ErrMismathedOrderTime = errors.New("orders time cannot actually be matched")
)

func (d *defaultCheck) query(def *OrderCheckerDef, token string, startTime int64, endTime int64) ([]map[string]interface{}, error) {
	queryParams := def.QueryParams
	for k, v := range queryParams {
		switch v {
		case `{trc20ContractAddress}`:
			v = Trc20ContractAddress
		case `{token}`:
			v = token
		case `{startTime}`:
			v = com.String(startTime)
		case `{endTime}`:
			v = com.String(endTime)
		}
		queryParams[k] = v
	}
	apiURL := def.BaseURL
	apiURL = strings.ReplaceAll(apiURL, `{token}`, token)
	req := d.client.R()
	if len(def.Headers) > 0 {
		req = req.SetHeaders(def.Headers)
	}
	resp, err := req.SetQueryParams(queryParams).Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf(`%w: %v`, ErrAPI, err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.New(resp.Status())
	}
	respData := param.Store{}
	err = json.Cjson.Unmarshal(resp.Body(), &respData)
	if err != nil {
		return nil, fmt.Errorf(`%w: %v`, ErrAPI, err)
	}
	if len(def.CountKeyName) > 0 {
		if def.GetInt64(respData, def.CountKeyName) <= 0 {
			return nil, nil
		}
	}
	list := def.Get(respData, def.ListKeyName)
	switch rows := list.(type) {
	case []interface{}:
		results := make([]map[string]interface{}, len(rows))
		for index, row := range rows {
			results[index], _ = row.(map[string]interface{})
		}
		return results, nil
	case []map[string]interface{}:
		return rows, nil
	default:
		return nil, fmt.Errorf(`%w: unsupported list data type: %T`, ErrAPI, list)
	}
}

func (d *defaultCheck) check(def *OrderCheckerDef, token string, startTime int64, endTime int64) error {
	rows, err := d.query(def, token, startTime, endTime)
	if err != nil {
		return err
	}
	for _, row := range rows {
		result := def.ParseResult(row)
		if result.ToToken != token || !result.IsSuccess(def) {
			result.Release()
			continue
		}
		amount, err := result.GetAmount(def)
		if err != nil {
			result.Release()
			return err
		}
		tradeId, err := data.GetTradeIdByWalletAddressAndAmount(token, amount)
		if err != nil {
			result.Release()
			return err
		}
		if len(tradeId) == 0 {
			result.Release()
			continue
		}
		order, err := data.GetOrderInfoByTradeId(tradeId)
		if err != nil {
			result.Release()
			return err
		}
		// 区块的确认时间必须在订单创建时间之后
		createTime := order.CreatedAt.Time.UnixMilli()
		if result.Timestamp < createTime {
			result.Release()
			return ErrMismathedOrderTime
		}
		// 到这一步就完全算是支付成功了
		req := &request.OrderProcessingRequest{
			Token:              token,
			TradeId:            tradeId,
			Amount:             amount,
			BlockTransactionId: result.TransactionId,
		}
		err = OrderProcessing(req)
		if err != nil {
			result.Release()
			return err
		}
		// 回调队列
		orderCallbackQueue, _ := handle.NewOrderCallbackQueue(order)
		mq.MClient.Enqueue(orderCallbackQueue, asynq.MaxRetry(5))
		// 发送机器人消息
		msgTpl := `
<b>📢📢有新的交易支付成功！</b>
<pre>交易号：%s</pre>
<pre>订单号：%s</pre>
<pre>请求支付金额：%f cny</pre>
<pre>实际支付金额：%f usdt</pre>
<pre>钱包地址：%s</pre>
<pre>订单创建时间：%s</pre>
<pre>支付成功时间：%s</pre>
`
		msg := fmt.Sprintf(msgTpl, order.TradeId, order.OrderId, order.Amount, order.ActualAmount, order.Token, order.CreatedAt.Time.Format(time.DateTime), time.Now().Format(time.DateTime))
		telegram.SendToBot(msg)
		result.Release()
	}
	return nil
}

type OrderCheckerDef struct {
	BaseURL          string            `yaml:"base_url"`
	QueryParams      map[string]string `yaml:"query_params"`
	Headers          map[string]string `yaml:"headers"`
	ListKeyName      string            `yaml:"list_key_name"`
	CountKeyName     string            `yaml:"count_key_name"`
	ItemKeyName      *ItemKeyName      `yaml:"item_key_name"`
	ItemSuccessValue string            `yaml:"item_success_value"`
	AmountDivisor    float64           `yaml:"amount_divisor"`
}

func (a *OrderCheckerDef) GetString(data param.Store, keyName string) string {
	parts := strings.Split(keyName, `.`)
	switch len(parts) {
	case 1:
		return data.String(parts[0])
	case 2:
		return data.GetStore(parts[0]).String(parts[len(parts)-1])
	default:
		return data.GetStoreByKeys(parts[0 : len(parts)-2]...).String(parts[len(parts)-1])
	}
}

func (a *OrderCheckerDef) Get(data param.Store, keyName string) interface{} {
	parts := strings.Split(keyName, `.`)
	switch len(parts) {
	case 1:
		return data.Get(parts[0])
	case 2:
		return data.GetStore(parts[0]).Get(parts[len(parts)-1])
	default:
		return data.GetStoreByKeys(parts[0 : len(parts)-2]...).Get(parts[len(parts)-1])
	}
}

func (a *OrderCheckerDef) GetInt64(data param.Store, keyName string) int64 {
	parts := strings.Split(keyName, `.`)
	switch len(parts) {
	case 1:
		return data.Int64(parts[0])
	case 2:
		return data.GetStore(parts[0]).Int64(parts[len(parts)-1])
	default:
		return data.GetStoreByKeys(parts[0 : len(parts)-2]...).Int64(parts[len(parts)-1])
	}
}

func (a *OrderCheckerDef) ParseResult(data param.Store) *Result {
	r := GetResult()
	r.ToToken = a.GetString(data, a.ItemKeyName.ToToken)
	r.Status = a.GetString(data, a.ItemKeyName.Status)
	r.Amount = a.GetString(data, a.ItemKeyName.Amount)
	r.Timestamp = a.GetInt64(data, a.ItemKeyName.Timestamp)
	r.TransactionId = a.GetString(data, a.ItemKeyName.TransactionId)
	return r
}

type ItemKeyName struct {
	ToToken       string `yaml:"to_token"`
	Status        string `yaml:"status"`
	Amount        string `yaml:"amount"`
	Timestamp     string `yaml:"timestamp"`
	TransactionId string `yaml:"transaction_id"`
}

var poolResult = sync.Pool{
	New: func() interface{} {
		//println(`-------------------> poolResultNew`)
		return &Result{}
	},
}

func GetResult() *Result {
	return poolResult.Get().(*Result)
}

func PutResult(r *Result) {
	r.Reset()
	poolResult.Put(r)
}

type Result struct {
	ToToken       string
	Status        string
	Amount        string
	Timestamp     int64
	TransactionId string
}

func (r *Result) IsSuccess(def *OrderCheckerDef) bool {
	return r.Status == def.ItemSuccessValue
}

func (r *Result) GetAmount(def *OrderCheckerDef) (float64, error) {
	var amount float64
	if def.AmountDivisor > 0 {
		decimalQuant, err := decimal.NewFromString(r.Amount)
		if err != nil {
			return amount, err
		}
		decimalDivisor := decimal.NewFromFloat(def.AmountDivisor)
		amount = decimalQuant.Div(decimalDivisor).InexactFloat64()
	} else {
		amount = param.AsFloat64(r.Amount)
	}
	return amount, nil
}

func (r *Result) Reset() {
	r.ToToken = ``
	r.Status = ``
	r.Amount = ``
	r.Timestamp = 0
	r.TransactionId = ``
}

func (r *Result) Release() {
	PutResult(r)
}
