package settings

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Conf 全局变量，保存viper从config.yaml文件读取出来的参数
var Conf = new(AppConfig)

type AppConfig struct {
	Port       int    `mapstructure:"port"`
	MaxMemory  int64  `mapstructure:"max_Memory"`
	Host       string `mapstructure:"host"`
	UploadPath string `mapstructure:"upload_path"`
}

// 加载配置文件config.yaml

func Init(filePath string) (err error) {
	// 使用命令行传入的参数
	viper.SetConfigFile(filePath)
	err = viper.ReadInConfig() // 查找并读取配置文件
	if err != nil {            // 处理读取配置文件的错误
		fmt.Printf("viper.ReadInConfig() failed, err:%v\n", err)
		return
	}
	// 把读取到的配置文件反序列化到 Conf 变量
	if err = viper.Unmarshal(Conf); err != nil {
		fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
	}
	// 监控config.yaml文件配置是否发生变化
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置文件修改了...")
		if err = viper.Unmarshal(Conf); err != nil {
			fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
		}
	})

	return
}
