package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/assimon/luuu/util/http_client"
	"github.com/spf13/viper"
)

var (
	AppDebug       bool
	MysqlDns       string
	RuntimePath    string
	LogSavePath    string
	StaticPath     string
	TgBotToken     string
	TgProxy        string
	TgManage       int64
	UsdtRate       float64
	CheckerDefPath string
	CheckerTimeout int64
	CheckerProxy   string
)

func Init() {
	viper.AddConfigPath("./")
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	gwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	AppDebug = viper.GetBool("app_debug")
	StaticPath = viper.GetString("static_path")
	RuntimePath = fmt.Sprintf(
		"%s%s",
		gwd,
		viper.GetString("runtime_root_path"))
	LogSavePath = fmt.Sprintf(
		"%s%s",
		RuntimePath,
		viper.GetString("log_save_path"))
	MysqlDns = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		url.QueryEscape(viper.GetString("mysql_user")),
		url.QueryEscape(viper.GetString("mysql_passwd")),
		fmt.Sprintf(
			"%s:%s",
			viper.GetString("mysql_host"),
			viper.GetString("mysql_port")),
		viper.GetString("mysql_database"))
	TgBotToken = viper.GetString("tg_bot_token")
	TgProxy = viper.GetString("tg_proxy")
	TgManage = viper.GetInt64("tg_manage")

	// - Added by admpub -

	CheckerDefPath = viper.GetString("checker_def_path")
	if len(CheckerDefPath) > 0 {
		CheckerDefPath = filepath.Join(gwd, CheckerDefPath)
	}
	CheckerTimeout = viper.GetInt64("checker_timeout")
	CheckerProxy = viper.GetString("checker_proxy")

	c := &Config{
		AppDebug:       AppDebug,
		MysqlDns:       MysqlDns,
		RuntimePath:    RuntimePath,
		LogSavePath:    LogSavePath,
		StaticPath:     StaticPath,
		TgBotToken:     TgBotToken,
		TgProxy:        TgProxy,
		TgManage:       TgManage,
		UsdtRate:       UsdtRate,
		CheckerDefPath: CheckerDefPath,
		CheckerTimeout: CheckerTimeout,
		CheckerProxy:   CheckerProxy,
	}
	if CheckerTimeout > 0 {
		http_client.TimeoutSeconds = CheckerTimeout
	}
	http_client.CheckerProxy = CheckerProxy
	err = FireInitialize(c)
	if err != nil {
		panic(err)
	}
}

func GetAppVersion() string {
	return "0.0.3"
}

func GetAppName() string {
	appName := viper.GetString("app_name")
	if len(appName) == 0 {
		return "epusdt"
	}
	return appName
}

func GetAppUri() string {
	return viper.GetString("app_uri")
}

func GetApiAuthToken() string {
	return viper.GetString("api_auth_token")
}

func GetUsdtRate() float64 {
	forcedUsdtRate := viper.GetFloat64("forced_usdt_rate")
	if forcedUsdtRate > 0 {
		return forcedUsdtRate
	}
	if UsdtRate <= 0 {
		return 6.4
	}
	return UsdtRate
}

func GetOrderExpirationTime() int {
	timer := viper.GetInt("order_expiration_time")
	if timer <= 0 {
		return 10
	}
	return timer
}

func GetOrderExpirationTimeDuration() time.Duration {
	timer := GetOrderExpirationTime()
	return time.Minute * time.Duration(timer)
}
