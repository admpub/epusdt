package http_client

import (
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	TimeoutSeconds int64 = 15
	CheckerProxy   string
)

// GetHttpClient 获取请求客户端
func GetHttpClient(proxys ...string) *resty.Client {
	client := resty.New()
	proxy := CheckerProxy
	// 如果有代理
	if len(proxys) > 0 {
		proxy = proxys[0]
	}
	if len(proxy) > 0 {
		client.SetProxy(proxy)
	}
	client.SetTimeout(time.Second * time.Duration(TimeoutSeconds))
	return client
}
