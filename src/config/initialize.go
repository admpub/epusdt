package config

import (
	"errors"

	"github.com/webx-top/echo"
)

type CoinTypes []*CoinType

var (
	ErrUnsupportedCurrency      = errors.New(`unsupported currency`)
	ErrUnsupportedCurrencyChain = errors.New(`unsupported currency chain`)
)

var (
	CurrencyChains = CoinTypes{
		&CoinType{
			Currency: `USDT`,
			Title:    `USDT`,
			ChainTypes: echo.KVList{
				echo.NewKV(`TRC20`, `TRC20`),
			},
		},
	}
)

func (c CoinTypes) Validate(currency string, chainType string) error {
	for _, cur := range c {
		if cur.Currency == currency {
			for _, cha := range cur.ChainTypes {
				if cha.K == chainType {
					return nil
				}
			}
			return ErrUnsupportedCurrencyChain
		}
	}
	return ErrUnsupportedCurrency
}

type CoinType struct {
	Currency   string      `yaml:"currency"`
	Title      string      `yaml:"title"`
	ChainTypes echo.KVList `yaml:"chain_types"`
}

type Config struct {
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
	CurrencyChains CoinTypes
}

var onInitializes = []func(*Config) error{}

func OnInitialize(fn func(*Config) error) {
	onInitializes = append(onInitializes, fn)
}

func FireInitialize(c *Config) error {
	for _, fn := range onInitializes {
		err := fn(c)
		if err != nil {
			return err
		}
	}
	return nil
}
