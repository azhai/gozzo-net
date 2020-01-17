package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/azhai/gozzo-net/cmd"
	"github.com/azhai/gozzo-net/network"
	"github.com/azhai/gozzo-net/unix"
)

var (
	conf          *cmd.RelaySetting
	app           *cmd.AppSection
	appname       string // 应用组名称
	filename      string // 配置文件路径
	port          uint
	number        int
	next          bool
	curr          bool
	relayServer   bool
	relay, server bool
	verbose       bool
)

// 解析参数
func init() {
	flag.StringVar(&appname, "a", "default", "应用组名称")
	flag.StringVar(&filename, "f", "servers.toml", "配置文件路径")

	// 端口，三选一
	flag.UintVar(&port, "p", 0, "server的端口")
	flag.IntVar(&number, "n", 0, "server端口的下标")
	// 比上面两个优先级高，且会更新配置文件
	flag.BoolVar(&next, "nx", true, "使用下一个端口")
	// 优先级最高，但不需要更新配置文件
	flag.BoolVar(&curr, "cr", false, "使用当前端口")

	// 两种服务，四种选择（还有rs=false都不运行）
	flag.BoolVar(&relay, "r", false, "只运行转发relay")
	flag.BoolVar(&server, "s", false, "只运行后端server")
	flag.BoolVar(&relayServer, "rs", false, "relay和server都运行")

	flag.BoolVar(&verbose, "v", false, "输出详细信息")
	flag.Parse()
}

// 创建代理
func run() {
	var err error
	// 获取应用配置
	if conf, err = cmd.GetConfig(filename); err != nil {
		if verbose {
			fmt.Println(err)
		}
		return
	}
	if app = conf.GetSection(appname); app == nil {
		if verbose {
			err = fmt.Errorf("Can not found the '%s' section", appname)
			fmt.Println(err)
		}
		return
	}
	// 确定后端server运行的端口
	var inPort uint16
	if port > 0 {
		inPort = uint16(port)
	} else {
		inPort = app.GetInPort(curr, next, number)
	}

	// 运行后端server，如果保存了前一个server的pid，先杀掉其进程
	if relayServer || server {
		cmd.KillProcess(app.Pid)
		strport := strconv.Itoa(int(inPort))
		go app.RunServer(strport, verbose)
	}
	// 保存配置，当前端口下标和进程pid已更新
	err = cmd.WriteConfig(filename, conf)
	if err != nil {
		if verbose {
			fmt.Println("write error: ", err)
		}
		return
	}
	// 运行端口转发的relay
	if relayServer || relay {
		addr := network.NewTCPAddr(app.Host, inPort)
		relayer := unix.NewRelayer(addr)
		proxy := unix.NewProxy("tcp", "", app.OutPort)
		events := network.Events{}
		events.Process = proxy.CreateProcess(relayer, unix.RelayData)
		proxy.Run(events)
	}
}

