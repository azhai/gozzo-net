// +build android darwin dragonfly freebsd linux netbsd openbsd solaris

package network

import (
	"net"
	"time"

	"github.com/felixge/tcpkeepalive"
)

// 为TCP连接设置保活时间
func (ka KeepAlive) ApplyTo(c *net.TCPConn) (err error) {
	if ka.Idle <= 0 {
		return
	}
	idle := time.Duration(ka.Idle) * time.Second
	// 总时间 = Idle + (Count - 1) * Interval
	interval := time.Duration(ka.Interval) * time.Second
	tcpkeepalive.SetKeepAlive(c, idle, ka.Count, interval)
	return
}
