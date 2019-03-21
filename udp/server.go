package udp

import (
	"net"
	"runtime"

	"github.com/azhai/gozzo-net/network"
)

// UDP服务器
type UDPServer struct {
	*network.Server
}

// 创建UDP服务器
func NewServer(server *network.Server) *UDPServer {
	return &UDPServer{Server: server}
}

// 服务启动阶段，执行Tick事件
func (s *UDPServer) Startup(events network.Events) (err error) {
	s.Trigger(events)
	return
}

// 服务停止阶段，关闭每一个网络连接
func (s *UDPServer) Shutdown(events network.Events) (err error) {
	s.Cleanup(func(c *network.Conn) error {
		return s.Finish(events, c)
	})
	return
}

// 开始服务，接受客户端连接，并进行读写
func (s *UDPServer) Run(events network.Events) (err error) {
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
	var (
		addr = network.GetUDPAddr(s.Address)
		conn *net.UDPConn
	)
	for {
		conn, err = net.ListenUDP("udp", addr)
		if err != nil {
			continue
		}
		c := network.NewUDPConn(conn)
		s.Execute(events, c)
	}
	return
}
