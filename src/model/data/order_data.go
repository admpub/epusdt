package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/assimon/luuu/model/dao"
	"github.com/assimon/luuu/model/mdb"
	"github.com/assimon/luuu/model/request"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var (
	CacheWalletAddressWithAmountToTradeIdKey = "wallet:%s_%v" // 钱包_待支付金额 : 交易号
)

// GetOrderInfoByOrderId 通过客户订单号查询订单
func GetOrderInfoByOrderId(orderId string) (*mdb.Orders, error) {
	order := new(mdb.Orders)
	err := dao.Mdb.Model(order).Limit(1).Find(order, "order_id = ?", orderId).Error
	return order, err
}

// GetOrderInfoByTradeId 通过交易号查询订单
func GetOrderInfoByTradeId(tradeId string) (*mdb.Orders, error) {
	order := new(mdb.Orders)
	err := dao.Mdb.Model(order).Limit(1).Find(order, "trade_id = ?", tradeId).Error
	return order, err
}

// CreateOrderWithTransaction 事务创建订单
func CreateOrderWithTransaction(tx *gorm.DB, order *mdb.Orders) error {
	err := tx.Model(order).Create(order).Error
	return err
}

// GetOrderByBlockIdWithTransaction 通过区块获取订单
func GetOrderByBlockIdWithTransaction(tx *gorm.DB, blockId string) (*mdb.Orders, error) {
	order := new(mdb.Orders)
	err := tx.Model(order).Limit(1).Find(order, "block_transaction_id = ?", blockId).Error
	return order, err
}

// OrderSuccessWithTransaction 事务支付成功
func OrderSuccessWithTransaction(tx *gorm.DB, req *request.OrderProcessingRequest) error {
	err := tx.Model(&mdb.Orders{}).Where("trade_id = ?", req.TradeId).Updates(map[string]interface{}{
		"block_transaction_id": req.BlockTransactionId,
		"status":               mdb.StatusPaySuccess,
		"callback_confirm":     mdb.CallBackConfirmNo,
	}).Error
	return err
}

// OrderReuseWithTransaction 复用以过期订单
func OrderReuseWithTransaction(tx *gorm.DB, tradeId string, update map[string]interface{}) error {
	err := tx.Model(&mdb.Orders{}).Where("trade_id = ?", tradeId).Updates(update).Error
	return err
}

// GetPendingCallbackOrders 查询出等待回调的订单
func GetPendingCallbackOrders() ([]mdb.Orders, error) {
	var orders []mdb.Orders
	err := dao.Mdb.Model(orders).
		Where("callback_num < ?", 5).
		Where("callback_confirm = ?", mdb.CallBackConfirmNo).
		Where("status = ?", mdb.StatusPaySuccess).
		Find(&orders).Error
	return orders, err
}

// SaveCallBackOrdersResp 保存订单回调结果
func SaveCallBackOrdersResp(order *mdb.Orders) error {
	err := dao.Mdb.Model(order).Where("id = ?", order.ID).Updates(map[string]interface{}{
		"callback_num":     gorm.Expr("callback_num + ?", 1),
		"callback_confirm": order.CallBackConfirm,
	}).Error
	return err
}

// UpdateOrderIsExpirationById 通过id设置订单过期
func UpdateOrderIsExpirationById(id uint64) error {
	err := dao.Mdb.Model(mdb.Orders{}).Where("id = ? AND status = ?", id, mdb.StatusWaitPay).Update("status", mdb.StatusExpired).Error
	return err
}

// ExistsOrderIsWaitPay 查询是否存在待支付订单
func ExistsOrderIsWaitPay() (bool, error) {
	order := new(mdb.Orders)
	err := dao.Mdb.Model(order).Select("id").Take(order, "status = ?", mdb.StatusWaitPay).Error
	if err == nil {
		return order.ID > 0, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

// GetTokenIsWaitPay 查询待支付订单token
func GetTokenIsWaitPay(token []string) ([]string, error) {
	orders := []*mdb.Orders{}
	err := dao.Mdb.Model(orders).Select("token").Group("token").Find(&orders, "status = ? AND token IN ?", mdb.StatusWaitPay, token).Error
	if err != nil {
		return nil, err
	}
	found := make([]string, len(orders))
	for i, v := range orders {
		found[i] = v.Token
	}
	return found, err
}

// GetTradeIdByWalletAddressAndAmount 通过钱包地址，支付金额获取交易号
func GetTradeIdByWalletAddressAndAmount(token string, amount float64) (string, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf(CacheWalletAddressWithAmountToTradeIdKey, token, amount)
	result, err := dao.Rdb.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return result, nil
}

// LockTransaction 锁定交易
func LockTransaction(token, tradeId string, amount float64, expirationTime time.Duration) error {
	ctx := context.Background()
	cacheKey := fmt.Sprintf(CacheWalletAddressWithAmountToTradeIdKey, token, amount)
	err := dao.Rdb.Set(ctx, cacheKey, tradeId, expirationTime).Err()
	return err
}

// UnLockTransaction 解锁交易
func UnLockTransaction(token string, amount float64) error {
	ctx := context.Background()
	cacheKey := fmt.Sprintf(CacheWalletAddressWithAmountToTradeIdKey, token, amount)
	err := dao.Rdb.Del(ctx, cacheKey).Err()
	return err
}
