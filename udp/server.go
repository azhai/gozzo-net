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
	s.Cleanup(events)
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
		sess := network.NewSession()
		c := network.NewUDPConn(conn, sess)
		if events.Opened != nil {
			if err = events.Opened(s.Server, c); err != nil {
				continue
			}
		}
		if events.Process != nil {
			go events.Process(s.Server, c)
		} else {
			go s.ProcessUDP(c, events)
		}
	}
	return
}

// 处理单个UDP连接
func (s *UDPServer) ProcessUDP(c *network.Conn, events network.Events) {
	defer s.CloseConn(c, events.Closed)
	// 下行阶段
	if events.Send != nil {
		c.ReadOnly = false
		go func(c *network.Conn) {
			for data := range c.Output {
				events.Send(c, data)
				runtime.Gosched()
			}
		}(c)
	}
	// 上行阶段
	if events.Prepare != nil && events.Receive != nil {
		sid := c.Session.GetId()
		spliter := events.Prepare(c, sid)
		if spliter == nil {
			return
		}
		datach := make(chan []byte)
		go func() {
			var saved bool
			for data := range datach {
				key := events.Receive(c, data, saved)
				if saved == false && key != "" {
					s.SaveConn(key, c)
					saved = true
				}
				runtime.Gosched()
			}
		}()
		c.ScanInput(datach, spliter)
	}
}
