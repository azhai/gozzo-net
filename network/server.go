package network

import (
	"net"
	"runtime"
	"sync"
	"time"
)

type CloseFunc func(c *Conn) error
type EachFunc func(k string, c *Conn) bool
type ProcessFunc func(s *Server, c *Conn)

// 事件集
type Events struct {
	Tick    func(t time.Time)
	Serving func(s *Server)
	Opened  func(s *Server, c *Conn) error
	Closed  func(s *Server, c *Conn, err error)
	Process ProcessFunc
	Prepare func(c *Conn, input chan<- []byte) error
	Receive func(c *Conn, data []byte, saved bool) string
	Send    func(c *Conn, data []byte)
}

// 网络连接集合
type Registry struct {
	conns sync.Map
}

// 遍历所有item
func (r *Registry) Each(op EachFunc) {
	r.conns.Range(func(key, value interface{}) bool {
		var next bool
		if c, ok := value.(*Conn); ok {
			next = op(key.(string), c)
		}
		return next // 为true时继续执行下一个，否则中断
	})
}

// 删除所有网络连接
func (r *Registry) Cleanup(closer CloseFunc) {
	r.Each(func(k string, c *Conn) bool {
		if closer != nil {
			closer(c)
			r.conns.Delete(k)
		} else {
			r.CloseConn(c, k)
		}
		return true // 继续执行下一个
	})
}

// 获取网络连接，key一般是设备（唯一）ID
func (r *Registry) LoadConn(key string) *Conn {
	if value, ok := r.conns.Load(key); ok {
		if c, ok := value.(*Conn); ok {
			return c
		}
	}
	return nil
}

// 保存网络连接，key一般是设备（唯一）ID
func (r *Registry) SaveConn(c *Conn, key string) bool {
	if key != "" {
		r.conns.Store(key, c)
		return true
	}
	return false
}

// 关闭网络连接，先执行Closed事件
func (r *Registry) CloseConn(c *Conn, key string) (err error) {
	if c != nil {
		if key != "" {
			r.conns.Delete(key)
		}
		err = c.Close()
	}
	return
}

// 同一个设备，旧连接将被新连接覆盖（会话ID不一样）
func (r *Registry) IsOverride(c *Conn, key string) bool {
	if sid := c.GetSessId(); sid != "" {
		if old := r.LoadConn(key); old != nil {
			return old.GetSessId() != sid
		}
	}
	return false
}

// 服务器
type Server struct {
	Address net.Addr
	Ticker  <-chan time.Time
	*Registry
}

// 创建服务器
func NewAddrServer(addr net.Addr) *Server {
	return &Server{Address: addr, Registry: new(Registry)}
}

// 创建TCP/UDP服务器
func NewPortServer(host string, port uint16) *Server {
	addr := NewTCPAddr(host, port)
	return NewAddrServer(addr)
}

// 创建unix socket服务器
func NewUnixServer(filename string) *Server {
	addr := NewUnixAddr(filename)
	return NewAddrServer(addr)
}

func (s *Server) SetTickInterval(secs int) {
	if secs > 0 {
		t := time.Duration(secs) * time.Second
		s.Ticker = time.Tick(t)
	}
}

// 执行Tick事件
func (s *Server) Trigger(events Events) {
	if events.Tick != nil && s.Ticker != nil {
		go func() {
			for t := range s.Ticker {
				events.Tick(t)
			}
		}()
	}
}

func (s *Server) Execute(events Events, c *Conn) {
	if events.Opened != nil {
		if err := events.Opened(s, c); err != nil {
			return
		}
	}
	go func() {
		defer s.Finish(events, c)
		if events.Process != nil {
			events.Process(s, c)
		} else {
			s.Process(events, c)
		}
	}()
}

// 关闭客户端
func (s *Server) Finish(events Events, c *Conn) error {
	if events.Closed != nil {
		events.Closed(s, c, c.LastError)
	}
	return s.CloseConn(c, c.GetSessId())
}

// 根据设备id下发数据
func (s *Server) SendTo(key string, data []byte) bool {
	if c := s.LoadConn(key); c != nil {
		if c.ReadOnly == false {
			c.Output <- data
		}
		return true
	}
	return false
}

// 处理单个连接
func (s *Server) Process(events Events, c *Conn) {
	// 下行阶段
	if events.Send != nil {
		c.ReadOnly = false
		go func(c *Conn) {
			for data := range c.Output {
				events.Send(c, data)
				runtime.Gosched()
			}
		}(c)
	}
	// 上行阶段
	if events.Prepare != nil && events.Receive != nil {
		datach := make(chan []byte)
		go func() {
			var saved bool
			for data := range datach {
				key := events.Receive(c, data, saved)
				if saved == false && key != "" {
					s.SaveConn(c, key)
					saved = true
				}
				runtime.Gosched()
			}
		}()
		c.LastError = events.Prepare(c, datach)
	}
}
