package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/azhai/gozzo-net/network"
	"github.com/azhai/gozzo-net/unix"
)

var (
	dhost string = "0.0.0.0"
	dport uint64 = 6379
	sport uint64 = 6380
)

// 解析参数，格式 6380:0.0.0.0:6379
func init() {
	if len(os.Args) <= 1 {
		return
	}
	pics := strings.SplitN(os.Args[1], ":", 3)
	if len(pics) != 3 {
		return
	}
	dhost = pics[1]
	dport, _ = strconv.ParseUint(pics[2], 10, 16)
	sport, _ = strconv.ParseUint(pics[0], 10, 16)
}

// 创建代理
func main() {
	proxy := unix.NewProxy("tcp", "", uint16(sport))
	addr := network.NewTCPAddr(dhost, uint16(dport))
	events := network.Events{}
	events.Process = proxy.CreateProcess(unix.NewRelayer(addr), unix.RelayData)
	proxy.Run("tcp", events)
}
