package route

import (
	"net/http"

	"github.com/assimon/luuu/controller/comm"
	"github.com/assimon/luuu/middleware"
	"github.com/labstack/echo/v4"
)

// RegisterRoute 路由注册
func RegisterRoute(e *echo.Echo) {
	e.Any("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello epusdt")
	})
	// ==== 支付相关=====
	payRoute := e.Group("/pay")
	// 收银台
	payRoute.GET("/checkout-counter/:trade_id", comm.Ctrl.CheckoutCounter)
	// 状态检测
	payRoute.GET("/check-status/:trade_id", comm.Ctrl.CheckStatus)

	apiV1Route := e.Group("/api/v1")
	// ====订单相关====
	orderRoute := apiV1Route.Group("/order", middleware.CheckApiSign())
	// 创建订单
	orderRoute.POST("/create-transaction", comm.Ctrl.CreateTransaction)
	orderRoute.POST("/query-transaction", comm.Ctrl.QueryTransaction)
}
