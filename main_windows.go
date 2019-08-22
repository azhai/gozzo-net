// +build windows

package main

import (
	"github.com/azhai/gozzo-utils/daemon"
)

func main() {
	name := "Gozzo_Redis_Proxy"
	desc := "Redis端口转发代理"
	daemon.WinMain(name, desc, run)
}
