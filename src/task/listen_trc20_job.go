package task

import (
	"sync"

	"github.com/assimon/luuu/model/data"
	"github.com/assimon/luuu/model/service"
	"github.com/assimon/luuu/util/log"
)

type ListenTrc20Job struct {
}

func (r ListenTrc20Job) Run() {
	walletAddress, err := data.GetAvailableWalletAddress()
	if err != nil {
		log.Sugar.Error(err)
		return
	}
	if len(walletAddress) <= 0 {
		return
	}
	tokens := make([]string, len(walletAddress))
	for i, v := range walletAddress {
		tokens[i] = v.Token
	}
	tokens, err = data.GetTokenIsWaitPay(tokens)
	if err != nil {
		log.Sugar.Error(err)
		return
	}
	if len(tokens) == 0 {
		return
	}
	var wg sync.WaitGroup
	if len(service.Defs()) > 0 {
		chr := service.Checker()
		for _, token := range tokens {
			wg.Add(1)
			go func(token string) {
				defer wg.Done()
				defer func() {
					if err := recover(); err != nil {
						log.Sugar.Error(err)
					}
				}()
				err := chr.Check(token)
				if err != nil {
					log.Sugar.Error(err)
				}
			}(token)
		}
		wg.Wait()
		return
	}
	for _, token := range tokens {
		wg.Add(1)
		go service.Trc20CallBack(token, &wg)
	}
	wg.Wait()
}
