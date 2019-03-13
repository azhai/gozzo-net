package network

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

// 检测TCP连接是否断开
// Idle: 没有数据往来（上行、下行都没有）多少秒后，发第一个检测包
// Count: 最大检测次数
// Interval: 相邻两个检测之间的时间间隔，单位：秒
type KeepAlive struct {
	Idle, Count, Interval int
}

type UDPOptions struct {
	ReadBuffer  int
	WriteBuffer int
	Deadline    int
}

// 设置UDP连接参数
func (opts UDPOptions) ApplyTo(c *net.UDPConn) (err error) {
	if opts.ReadBuffer > 0 {
		err = c.SetReadBuffer(opts.ReadBuffer)
	}
	if opts.WriteBuffer > 0 {
		err = c.SetWriteBuffer(opts.WriteBuffer)
	}
	if opts.Deadline > 0 {
		secs := time.Duration(opts.Deadline)
		err = c.SetDeadline(time.Now().Add(secs * time.Second))
	}
	return
}

type TCPOptions struct {
	NoDelay bool // default true
	Linger  int  // default -1
	KeepAlive
	UDPOptions
}

var DefaultTCPOptions = TCPOptions{
	NoDelay: true,
	Linger:  -1,
}

// 设置TCP连接参数
func (opts TCPOptions) ApplyTo(c *net.TCPConn) (err error) {
	// 与UDP相同的参数
	if opts.ReadBuffer > 0 {
		err = c.SetReadBuffer(opts.ReadBuffer)
	}
	if opts.WriteBuffer > 0 {
		err = c.SetWriteBuffer(opts.WriteBuffer)
	}
	if opts.Deadline > 0 {
		secs := time.Duration(opts.Deadline)
		err = c.SetDeadline(time.Now().Add(secs * time.Second))
	}
	// TCP独有的参数
	err = opts.KeepAlive.ApplyTo(c)
	err = c.SetNoDelay(opts.NoDelay)
	err = c.SetLinger(opts.Linger)
	return
}

type Conn struct {
	kind    string
	conn    net.Conn
	reader  *bufio.Reader
	Session *Session
	// 下发指令专用，回复请直接调用Write
	Output    chan []byte
	ReadOnly  bool
	IsActive  bool
	LastError error
}

func newConn(kind string, conn net.Conn, isActive bool, sess *Session) *Conn {
	return &Conn{
		kind:      kind,
		conn:      conn,
		reader:    nil,
		Session:   sess,
		Output:    make(chan []byte),
		ReadOnly:  true,
		IsActive:  isActive,
		LastError: nil,
	}
}

func NewTCPConn(conn *net.TCPConn, sess *Session) *Conn {
	return newConn("tcp", conn, conn != nil, sess)
}

func NewUDPConn(conn *net.UDPConn, sess *Session) *Conn {
	return newConn("udp", conn, conn != nil, sess)
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

// 返回原始的网络连接
// 注意：调用过Peek()操作后，原始连接不能再用于读（会丢失前n个字节）
func (c *Conn) GetRawConn() net.Conn {
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

// 扫描和分割输入数据
func (c *Conn) ScanInput(datach chan<- []byte, spliter bufio.SplitFunc) {
	defer func() { // 记录错误或异常
		err, ok := recover().(error)
		if ok && err != nil {
			c.LastError = err
		}
	}()
	// 扫描输入
	scanner := bufio.NewScanner(c.GetReader())
	scanner.Split(spliter)
	for scanner.Scan() {
		body := scanner.Bytes()
		if n := len(body); n > 0 {
			datach <- body
		}
	}
	c.LastError = scanner.Err()
}
