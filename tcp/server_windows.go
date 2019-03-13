// +build windows

package tcp

import "net"

// 监听TCP端口
func ListenTCP(address string) (net.Listener, error) {
	return net.Listen("tcp", address)
}
