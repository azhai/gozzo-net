package network

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

type IPckConn interface {
	SetReadBuffer(bytes int) error
	SetWriteBuffer(bytes int) error
	net.Conn
}

type Options struct {
	ReadBuffer  int
	WriteBuffer int
	Deadline    int
}

// 设置UDP连接参数
func (opts Options) ApplyConn(c IPckConn) (err error) {
	if opts.ReadBuffer > 0 {
		err = c.SetReadBuffer(opts.ReadBuffer)
	}
	if err == nil && opts.WriteBuffer > 0 {
		err = c.SetWriteBuffer(opts.WriteBuffer)
	}
	if err == nil && opts.Deadline > 0 {
		secs := time.Duration(opts.Deadline)
		err = c.SetDeadline(time.Now().Add(secs * time.Second))
	}
	return
}

// 检测TCP连接是否断开
// Idle: 没有数据往来（上行、下行都没有）多少秒后，发第一个检测包
// Count: 最大检测次数
// Interval: 相邻两个检测之间的时间间隔，单位：秒
type KeepAlive struct {
	Idle, Count, Interval int
}

type TCPOptions struct {
	NoDelay bool // default true
	Linger  int  // default -1
	KeepAlive
	Options
}

var DefaultTCPOptions = TCPOptions{
	NoDelay: true,
	Linger:  -1,
}

// 设置TCP连接参数
func (opts TCPOptions) ApplyTCP(c *net.TCPConn) (err error) {
	err = opts.ApplyConn(c)
	// TCP独有的参数
	if err == nil {
		err = opts.KeepAlive.ApplyTo(c)
	}
	if err == nil {
		err = c.SetNoDelay(opts.NoDelay)
	}
	if err == nil {
		err = c.SetLinger(opts.Linger)
	}
	return
}

type Conn struct {
	kind    string
	conn    IPckConn
	reader  *bufio.Reader
	Session *Session
	// 下发指令专用，回复请直接调用Write
	Output    chan []byte
	ReadOnly  bool
	IsActive  bool
	LastError error
}

func newConn(kind string, conn IPckConn, isActive bool) *Conn {
	return &Conn{
		kind:     kind,
		conn:     conn,
		Output:   make(chan []byte),
		ReadOnly: true,
		IsActive: isActive,
	}
}

func NewTCPConn(conn *net.TCPConn) *Conn {
	return newConn("tcp", conn, conn != nil)
}

func NewUDPConn(conn *net.UDPConn) *Conn {
	return newConn("udp", conn, conn != nil)
}

func NewUnixConn(conn *net.UnixConn) *Conn {
	return newConn("unix", conn, conn != nil)
}

func (c *Conn) Close() error {
	if c.IsActive {
		c.IsActive = false
		c.ReadOnly = false
		close(c.Output)
		return c.conn.Close()
	}
	return nil
}

func (c *Conn) GetKind() string {
	return c.kind
}

func (c *Conn) GetSessId() string {
	if sess := c.Session; sess != nil {
		return sess.GetId()
	}
	return ""
}

// 返回原始的网络连接
// 注意：调用过Peek()操作后，原始连接不能再用于读（会丢失前n个字节）
func (c *Conn) GetRawConn() IPckConn {
	return c.conn
}

func (c *Conn) GetReader() *bufio.Reader {
	if c.reader == nil && c.IsActive {
		c.reader = bufio.NewReader(c.conn)
	}
	return c.reader
}

// 往前读n个字节，但不移动游标
// 注意：原始conn的读游标会向前移动，所有读的地方，用GetReader()代替GetConn()
func (c *Conn) Peek(n int) ([]byte, error) {
	return c.GetReader().Peek(n)
}

// 快速发送数据，但连接断开也不会去重连
func (c *Conn) QuickSend(data []byte) error {
	if c.IsActive == false {
		return fmt.Errorf("Lost connection")
	}
	sent, err := c.conn.Write(data)
	if err == nil && sent < len(data) {
		return fmt.Errorf("Only sent %d bytes", sent)
	}
	return err
}
