package dao

import (
	stdLog "log"
	"os"
	"time"

	"github.com/assimon/luuu/config"
	"github.com/assimon/luuu/util/log"
	"github.com/gookit/color"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var Mdb *gorm.DB

// MysqlInit 数据库初始化
func MysqlInit() {
	var err error
	Mdb, err = gorm.Open(mysql.Open(config.MysqlDns), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   viper.GetString("mysql_table_prefix"),
			SingularTable: true,
		},
		Logger: logger.New(stdLog.New(os.Stdout, "\r\n", stdLog.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		}),
	})
	if err != nil {
		panic(err)
	}
	if config.AppDebug {
		Mdb = Mdb.Debug()
	}
	sqlDB, err := Mdb.DB()
	if err != nil {
		color.Red.Printf("[store_db] mysql get DB,err=%s\n", err)
		panic(err)
	}
	sqlDB.SetMaxIdleConns(viper.GetInt("mysql_max_idle_conns"))
	sqlDB.SetMaxOpenConns(viper.GetInt("mysql_max_open_conns"))
	sqlDB.SetConnMaxLifetime(time.Hour * time.Duration(viper.GetInt("mysql_max_life_time")))
	err = sqlDB.Ping()
	if err != nil {
		color.Red.Printf("[store_db] mysql connDB err:%s", err.Error())
		panic(err)
	}
	log.Sugar.Debug("[store_db] mysql connDB success")
}
