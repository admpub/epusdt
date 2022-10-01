package config

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
