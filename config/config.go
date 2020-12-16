/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/17/20 9:30 AM
* @Description: 全局配置单例
***********************************************************************/

// TODO 暂时不考虑热更新

package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	cfg     *TomlConfig
	once    sync.Once
	cfgLock sync.RWMutex
)

// Init 初始化
func Init(cfgPath string) {
	once.Do(func() {
		// 第一次加载配置
		reloadConfig(cfgPath)

		// 自定义信号，收到信号则重载配置
		s := make(chan os.Signal, 1)
		signal.Notify(s, syscall.SIGUSR1)
		go func() {
			for {
				<-s
				reloadConfig(cfgPath)
				log.Println("Reloaded config")
			}
		}()
	})
}

// Global 全局单例
func Global() *TomlConfig {
	cfgLock.RLock()
	defer cfgLock.RUnlock()
	return cfg
}

// reloadConfig 加载/重加载配置
func reloadConfig(cfgPath string) {
	//filePath, err := filepath.Abs("./ch3/config.toml")
	//if err != nil {
	//	panic(err)
	//}
	fmt.Printf("parse toml file once. filePath: %s\n", cfgPath)
	conf := new(TomlConfig)
	if _, err := toml.DecodeFile(cfgPath, conf); err != nil {
		panic(err)
	}
	cfgLock.Lock()
	defer cfgLock.Unlock()
	cfg = conf
}

////////////////////// 单纯解析一个配置 ////////////////////////

// ParseConfig 以toml协议解析配置
func ParseConfig(r io.Reader) (*TomlConfig, error) {
	c := new(TomlConfig)
	if _, err := toml.DecodeReader(r, c); err != nil {
		return nil, err
	}
	return c, nil
}
