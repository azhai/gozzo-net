// +build windows

package network

import (
	"net"
	"time"
)

// 为TCP连接设置保活时间
func (ka KeepAlive) ApplyTo(c *net.TCPConn) (err error) {
	if ka.Idle <= 0 {
		return
	}
	idle := time.Duration(ka.Idle) * time.Second
	if ka.Count <= 0 || ka.Interval <= 0 {
		err = c.SetKeepAlive(true)
		err = c.SetKeepAlivePeriod(idle)
	}
	return
}
