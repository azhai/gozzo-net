package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/azhai/gozzo-utils/filesystem"
	"github.com/azhai/gozzo-utils/logging"
	"github.com/azhai/gozzo-utils/queue"
)

/**
***********************************************************
* settings 配置文件解析
***********************************************************
**/

var (
	channel  *queue.Channel
)

type ServSetting struct {
	Log    LogSection
	Server ServSection
	Proto  BuffSection
	Rabbit MesgSection
}

// 日志配置
type LogSection struct {
	Level  string
	Logdir string
}

func (c LogSection) GetLogger() *logging.Logger {
	return logging.NewLogger(c.Level, c.Logdir)
}

// 读缓冲配置
type BuffSection struct {
	ReadBuffSize int // 读缓冲大小（字节）
}

func (c BuffSection) GetBuffSize() int {
	if c.ReadBuffSize > 0 {
		return c.ReadBuffSize
	}
	return 4096
}

// 服务端配置
type ServSection struct {
	Host string
	Port int
	Tick int // 打点器间隔（秒）
	BuffSection
}

// 队列配置
type MesgSection struct {
	Url      string
	Exchange string
	Routings []string
}

func (c MesgSection) Push(id int, msg *queue.Message) error {
	if channel == nil {
		channel = queue.NewChannel(c.Url)
	}
	route := c.Routings[id%len(c.Routings)]
	return channel.PushMessage(msg, route, c.Exchange)
}

/**
***********************************************************
* servers 配置文件解析
***********************************************************
**/

// 配置，含多个应用配置
type RelaySetting map[string]*AppSection

// 应用配置
type AppSection struct {
	Host    string
	Ports   []uint16
	OutPort uint16 `toml:outport`
	Prog    string
	Args    string
	Curr    int
	Pid     int
}

// 获取其中一个应用的配置
func (conf RelaySetting) GetSection(name string) *AppSection {
	if c, ok := conf[name]; ok {
		return c
	}
	return nil
}

// 计算端口，优先使用next即下一个，number为指定端口下标
func (app *AppSection) GetInPort(curr, next bool, number int) uint16 {
	var size int
	if size = len(app.Ports); size == 0 {
		return 0
	}
	if curr {
		number = app.Curr % size
	} else if next {
		number = (app.Curr + 1) % size
		app.Curr = number // 记录下来
	} else {
		number = number % size
	}
	return app.Ports[number]
}

// 运行后端服务
func (app *AppSection) RunServer(port string, verbose bool) int {
	dir, _ := filepath.Abs(filesystem.GetRunDir())
	prog := strings.ReplaceAll(app.Prog, "$dir", dir)
	rpl := strings.NewReplacer("$port", port, "$host", app.Host)
	args := strings.Fields(rpl.Replace(app.Args))
	args = append(args, "> /dev/null", "&!")

	// Command不允许将多个参数加空格作为一个参数用
	command := exec.Command(prog, args...)
	fmt.Println(command)
	out, err := command.Output()
	if err != nil {
		if verbose {
			fmt.Println("exec error: ", err)
		}
		return 0
	}
	if runtime.GOOS == "linux" { // 从输出中解析pid，仅限Linux
		fields := strings.Fields(string(out))
		app.Pid, _ = strconv.Atoi(fields[len(fields)-1])
		if verbose {
			fmt.Printf("The pid = %d\n", app.Pid)
		}
	}
	return app.Pid
}

// 解析配置
func GetConfig(filename string) (*RelaySetting, error) {
	var conf = new(RelaySetting)
	if _, exists := filesystem.FileSize(filename); !exists {
		return conf, fmt.Errorf("File %s is not exists !", filename)
	}
	fullPath := filesystem.GetAbsFile(filename)
	_, err := toml.DecodeFile(fullPath, &conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

// 写入配置
func WriteConfig(filename string, conf *RelaySetting) error {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(conf)
	if err == nil {
		err = ioutil.WriteFile(filename, buf.Bytes(), 0644)
	}
	return err
}

// 根据pid杀进程
func KillProcess(pid int) error {
	if pid <= 0 {
		return nil
	}
	proc, err := os.FindProcess(pid)
	if err == nil {
		err = proc.Kill()
	}
	return err
}