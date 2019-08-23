package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/azhai/gozzo-utils/logging"
	"github.com/azhai/gozzo-utils/metrics"
	"github.com/azhai/gozzo-utils/queue"
)

var (
	filename = "settings.toml"
	conf     appConfig       // 配置
	logger   *logging.Logger // 日志
	reporter *metrics.DummyReporter
	channel  *queue.Channel
)

/**
***********************************************************
* 配置文件解析
***********************************************************
**/

type appConfig struct {
	Log    logConfig
	Server srvConfig
	Proto  buffConfig
	Rabbit mqConfig
}

// 日志配置
type logConfig struct {
	Level  string
	Logdir string
}

func (c logConfig) GetLogger() *logging.Logger {
	return logging.NewLogger(c.Level, c.Logdir)
}

// 读缓冲配置
type buffConfig struct {
	ReadBuffSize int // 读缓冲大小（字节）
}

func (c buffConfig) GetBuffSize() int {
	if c.ReadBuffSize > 0 {
		return c.ReadBuffSize
	}
	return 4096
}

// 服务端配置
type srvConfig struct {
	Host string
	Port int
	Tick int // 打点器间隔（秒）
	buffConfig
}

// 队列配置
type mqConfig struct {
	Url      string
	Exchange string
	Routings []string
}

func (c mqConfig) Push(id int, msg *queue.Message) error {
	if channel == nil {
		channel = queue.NewChannel(c.Url)
	}
	route := c.Routings[id%len(c.Routings)]
	return channel.PushMessage(c.Exchange, route, msg)
}

func init() {
	// 解析配置和创建日志
	if flag.NArg() > 0 {
		filename = flag.Arg(0)
	} else {
		filename = GetAbsFile(filename)
	}
	_, err := toml.DecodeFile(filename, &conf)

	if CheckError(err) {
		logger = conf.Log.GetLogger()
	}
	var names = []string{"opened", "closed", "received"}
	reporter = metrics.NewDummyReporter(names)
}

/**
***********************************************************
* 辅助函数
***********************************************************
**/

// 遇到错误时记录到日志
func CheckError(err error) bool {
	if err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		return false
	}
	return true
}

// 取得文件的绝对路径
func GetAbsFile(fname string) string {
	if filepath.IsAbs(fname) == false {
		// 相对于程序运行目录
		origDir := filepath.Dir(os.Args[0])
		dir, err := filepath.Abs(origDir)
		if err != nil {
			return ""
		}
		dir = strings.Replace(dir, "\\", "/", -1)
		fname = filepath.Join(dir, fname)
	}
	return fname
}
