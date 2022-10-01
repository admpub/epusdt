package task

import (
	"os"
	"sync"

	"github.com/assimon/luuu/config"
	"github.com/assimon/luuu/model/data"
	"github.com/assimon/luuu/model/service"
	"github.com/assimon/luuu/util/log"
	"gopkg.in/yaml.v3"
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
	if len(config.CheckerDefPath) > 0 {
		var defs []*service.OrderCheckerDef
		var b []byte
		b, err = os.ReadFile(config.CheckerDefPath)
		if err == nil {
			err = yaml.Unmarshal(b, &defs)
		}
		if err == nil {
			chr := service.Checker()
			for _, address := range walletAddress {
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
				}(address.Token)
			}
			wg.Wait()
			return
		}

		log.Sugar.Error(err)
	}
	for _, address := range walletAddress {
		wg.Add(1)
		go service.Trc20CallBack(address.Token, &wg)
	}
	wg.Wait()
}
