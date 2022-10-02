package task

import (
	"sync"

	"github.com/assimon/luuu/model/data"
	"github.com/assimon/luuu/model/service"
	"github.com/assimon/luuu/util/log"
)

type ListenTrc20Job struct {
}

var gListenTrc20JobLock sync.Mutex

func (r ListenTrc20Job) Run() {
	gListenTrc20JobLock.Lock()
	defer gListenTrc20JobLock.Unlock()
	walletAddress, err := data.GetAvailableWalletAddress()
	if err != nil {
		log.Sugar.Error(err)
		return
	}
	if len(walletAddress) <= 0 {
		return
	}
	var wg sync.WaitGroup
	if len(service.Defs()) > 0 {
		for _, address := range walletAddress {
			wg.Add(1)
			if len(address.Currency) == 0 {
				address.Currency = service.DefaultCurrency
			}
			if len(address.ChainType) == 0 {
				address.ChainType = service.DefaultChainType
			}
			go func(token string, currency string, chainType string) {
				defer wg.Done()
				defer func() {
					if err := recover(); err != nil {
						log.Sugar.Error(err)
					}
				}()
				chr := service.Checker(currency, chainType)
				if chr == nil {
					log.Sugar.Errorf(`unsupported checker for currency-chain: %s-%s`, currency, chainType)
					return
				}
				err := chr.Check(token, currency, chainType)
				if err != nil {
					log.Sugar.Error(err)
				}
			}(address.Token, address.Currency, address.ChainType)
		}
		wg.Wait()
		return
	}
	for _, address := range walletAddress {
		wg.Add(1)
		if len(address.Currency) == 0 {
			address.Currency = service.DefaultCurrency
		}
		if len(address.ChainType) == 0 {
			address.ChainType = service.DefaultChainType
		}
		go service.Trc20CallBack(address.Token, address.Currency, address.ChainType, &wg)
	}
	wg.Wait()
}
