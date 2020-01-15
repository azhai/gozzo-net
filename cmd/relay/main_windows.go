// +build windows

package main

import (
	"github.com/azhai/gozzo-utils/daemon"
)

func main() {
	name := "API Switcher"
	desc := "后端API轮换代理"
	daemon.WinMain(name, desc, run)
}
