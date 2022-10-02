package data

import (
	"fmt"
	"testing"

	"github.com/assimon/luuu/config"
	"github.com/assimon/luuu/model/dao"
	"github.com/assimon/luuu/util/log"
	"github.com/spf13/viper"
	"github.com/webx-top/com"
)

var token = `123456789009876`

func TestMain(m *testing.M) {
	log.Init()
	config.MysqlDns = `root:root@tcp(127.0.0.1:3306)/equsdt_test?charset=utf8mb4&parseTime=True&loc=Local`
	viper.Set(`mysql_table_prefix`, ``)
	viper.Set(`mysql_max_idle_conns`, `10`)
	viper.Set(`mysql_max_open_conns`, `100`)
	viper.Set(`mysql_max_life_time`, `6`)
	dao.MysqlInit()
	dao.Mdb = dao.Mdb.Debug()
	m.Run()
}

func TestAddWalletAddress(t *testing.T) {
	w, err := AddWalletAddress(token, `USDT`, `TRC20`)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	//defer DeleteWalletAddressById(w.ID)
	fmt.Printf("new wallet address: %v\n", com.Dump(w, false))
}

func TestGetAvailableWalletAddress(t *testing.T) {
	rows, err := GetAvailableWalletAddress()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	fmt.Printf("get wallet addresses: %v\n", com.Dump(rows, false))
}

func TestGetWalletAddressByToken(t *testing.T) {
	row, err := GetWalletAddressByToken(token)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	com.Dump(row)
	if row.ID <= 0 {
		panic(`not found`)
	}
	err = DeleteWalletAddressById(row.ID)
	if err != nil {
		panic(err)
	}
}
