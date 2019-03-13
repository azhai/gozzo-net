package network

import (
	"bufio"
	"net"
	"sync"
	"time"
)

type PreCloseConn func(r *Registry, c *Conn, err error)

// 事件集
type Events struct {
	Tick    func(t time.Time)
	Closed  PreCloseConn
	Serving func(s *Server)
	Opened  func(s *Server, c *Conn) error
	Process func(s *Server, c *Conn)
	Prepare func(c *Conn, sid string) bufio.SplitFunc
	Receive func(c *Conn, data []byte, saved bool) string
	Send    func(c *Conn, data []byte)
}

// 网络连接集合
type Registry struct {
	conns sync.Map
}

// 删除所有网络连接
func (r Registry) Cleanup(events Events) {
	r.conns.Range(func(key, value interface{}) bool {
		if c, ok := value.(*Conn); ok {
			r.CloseConn(c, events.Closed)
			r.conns.Delete(key.(string))
		}
		return true // 继续执行下一个
	})
}

// 关闭网络连接，先执行Closed事件
func (r Registry) CloseConn(c *Conn, pcc PreCloseConn) (err error) {
	if c != nil {
		if pcc != nil {
			pcc(&r, c, c.LastError)
		}
		err = c.Close()
	}
	return
}

// 获取网络连接，key一般是设备（唯一）ID
func (r Registry) LoadConn(key string) *Conn {
	if value, ok := r.conns.Load(key); ok {
		if c, ok := value.(*Conn); ok {
			return c
		}
	}
	return nil
}

// 保存网络连接，key一般是设备（唯一）ID
func (r Registry) SaveConn(key string, c *Conn) bool {
	if key == "" {
		return false
	}
	r.conns.Store(key, c)
	return true
}

// 同一个设备，旧连接将被新连接覆盖（会话ID不一样）
func (r Registry) IsOverride(key string, c *Conn) bool {
	if old := r.LoadConn(key); old != nil {
		sid := c.Session.GetId()
		return old.Session.GetId() != sid
	}
	return false
}

// 服务器
type Server struct {
	Address  net.Addr
	TickMsec int
	Registry
}

// 创建TCP服务器
func NewServer(host string, port uint16, tick int) *Server {
	addr, _ := NewTCPAddr(host, port)
	return &Server{Address: addr, TickMsec: tick}
}

// 创建TCP服务器
func NewAddrServer(addr net.Addr, tick int) *Server {
	return &Server{Address: addr, TickMsec: tick}
}

// 执行Tick事件
func (s *Server) Trigger(events Events) {
	if events.Tick != nil && s.TickMsec > 0 {
		msecs := time.Duration(s.TickMsec)
		go func(msecs time.Duration) {
			ticker := time.Tick(msecs * time.Millisecond)
			for t := range ticker {
				events.Tick(t)
			}
		}(msecs)
	}
}

func (s *Server) SendTo(key string, data []byte) bool {
	if c := s.LoadConn(key); c != nil {
		if c.ReadOnly == false {
			c.Output <- data
		}
		return true
	}
	return false
}
