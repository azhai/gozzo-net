package tcp

import (
	"net"
	"runtime"

	"github.com/azhai/gozzo-net/network"
)

// TCP服务器
type TCPServer struct {
	listener *net.TCPListener
	*network.Server
}

// 创建TCP服务器
func NewServer(server *network.Server) *TCPServer {
	return &TCPServer{Server: server}
}

// 服务启动阶段，执行Tick事件
func (s *TCPServer) Startup(events network.Events) (err error) {
	addr := network.GetTCPAddr(s.Address)
	s.listener, err = ListenTCP(addr.String())
	if err == nil {
		s.Trigger(events)
	}
	return
}

// 服务停止阶段，关闭每一个网络连接
func (s *TCPServer) Shutdown(events network.Events) (err error) {
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
func (s *TCPServer) Run(events network.Events) (err error) {
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
	var conn *net.TCPConn
	for {
		conn, err = s.listener.AcceptTCP()
		if err != nil {
			continue
		}
		c := network.NewTCPConn(conn)
		s.Execute(events, c)
	}
	return
}
