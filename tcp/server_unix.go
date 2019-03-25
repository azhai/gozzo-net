// +build android darwin dragonfly freebsd linux netbsd openbsd solaris

package tcp

import (
	"net"

	"github.com/kavu/go_reuseport"
)

// 监听TCP端口
func ListenTCP(address string) (*net.TCPListener, error) {
	listener, err := reuseport.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return listener.(*net.TCPListener), err
}
