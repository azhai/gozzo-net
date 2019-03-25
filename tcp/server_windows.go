// +build windows

package tcp

import "net"

// 监听TCP端口
func ListenTCP(address string) (*net.TCPListener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return listener.(*net.TCPListener), err
}
