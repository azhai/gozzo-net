package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/azhai/gozzo-utils/filesystem"
)

// 配置，含多个应用配置
type Config map[string]*AppConfig

// 应用配置
type AppConfig struct {
	Host    string
	Ports   []uint16
	OutPort uint16 `toml:outport`
	Prog    string
	Args    string
	Curr    int
	Pid     int
}

// 获取其中一个应用的配置
func (conf Config) GetSection(name string) *AppConfig {
	if c, ok := conf[name]; ok {
		return c
	}
	return nil
}

// 计算端口，优先使用next即下一个，number为指定端口下标
func (app *AppConfig) GetInPort(curr, next bool, number int) uint16 {
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
func (app *AppConfig) RunServer(port string, verbose bool) int {
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
func GetConfig(filename string) (*Config, error) {
	var conf = new(Config)
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
func WriteConfig(filename string, conf *Config) error {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(conf)
	if err == nil {
		err = ioutil.WriteFile(filename, buf.Bytes(), 0644)
	}
	return err
}
