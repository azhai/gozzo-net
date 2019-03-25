// +build android darwin dragonfly freebsd linux netbsd openbsd solaris

package unix

import (
	"fmt"
	"net"
	"runtime"

	"github.com/azhai/gozzo-net/network"
)

// Unix socket 服务器
type UnixServer struct {
	listener *net.UnixListener
	*network.Server
}

// 创建Unix服务器
func NewServer(server *network.Server) *UnixServer {
	return &UnixServer{Server: server}
}

// 服务启动阶段，执行Tick事件
func (s *UnixServer) Startup(events network.Events) (err error) {
	if addr, ok := s.Address.(*net.UnixAddr); ok {
		s.listener, err = net.ListenUnix("unix", addr)
	} else {
		err = fmt.Errorf("The address is not a UnixAddr object")
	}
	if err == nil {
		s.Trigger(events)
	}
	return
}

// 服务停止阶段，关闭每一个网络连接
func (s *UnixServer) Shutdown(events network.Events) (err error) {
	if s.listener != nil {
		if err = s.listener.Close(); err != nil {
			return
		}
	}
	s.Cleanup(func(c *network.Conn) error {
		return s.Finish(events, c)
	})
	return
}

// 开始服务，接受客户端连接，并进行读写
func (s *UnixServer) Run(events network.Events) (err error) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// 启动与停止
	if err = s.Startup(events); err != nil {
		return
	}
	defer s.Shutdown(events)
	if events.Serving != nil {
		events.Serving(s.Server)
	}
	// 循环接收和处理连接
	var conn *net.UnixConn
	for {
		conn, err = s.listener.AcceptUnix()
		if err != nil {
			continue
		}
		c := network.NewUnixConn(conn)
		s.Execute(events, c)
	}
	return
}
