package task

import (
	"sync"

	"github.com/assimon/luuu/model/data"
	"github.com/assimon/luuu/util/log"
	"github.com/robfig/cron/v3"
)

func Start() {
	c := cron.New()
	// 汇率监听
	c.AddJob("@every 60s", UsdtRateJob{})
	// trc20钱包监听
	c.AddJob("@every 5s", NewWalletListener(ListenTrc20Job{}))
	c.Start()
}

func NewWalletListener(job cron.Job) cron.Job {
	return &listenWallet{
		job: job,
		mu:  sync.RWMutex{},
	}
}

type listenWallet struct {
	job cron.Job
	mu  sync.RWMutex
}

func (r *listenWallet) Run() {
	r.mu.Lock()
	defer r.mu.Unlock()
	exists, err := data.ExistsOrderIsWaitPay()
	if err != nil {
		log.Sugar.Error(err)
		return
	}
	if exists {
		return
	}
	r.job.Run()
}
