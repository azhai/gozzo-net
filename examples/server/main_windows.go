// +build windows

package main

import (
	"github.com/azhai/gozzo-utils/daemon"
)

func main() {
	name := "Test_TCP_Server"
	desc := "TCP数据记录服务"
	daemon.WinMain(name, desc, run)
}
