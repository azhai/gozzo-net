package tcp

import (
	"github.com/azhai/gozzo-net/network"
)

// TCP 客户端
type TCPClient struct {
	options  network.TCPOptions
	dialplan *network.DialPlan
	Conn     *network.Conn
}

// 创建TCP客户端
func NewClient(plan *network.DialPlan, opts network.TCPOptions) *TCPClient {
	return &TCPClient{dialplan: plan, options: opts}
}

func (c *TCPClient) Close() error {
	if c.Conn == nil {
		return nil
	}
	return c.Conn.Close()
}

func (c *TCPClient) GetPlan() *network.DialPlan {
	return c.dialplan
}

func (c *TCPClient) GetConn() *network.Conn {
	return c.Conn
}

func (c *TCPClient) SetConn(conn *network.Conn) {
	c.Conn = conn
}

func (c *TCPClient) Dialing() (*network.Conn, error) {
	conn, err := c.dialplan.DialTCP()
	if err == nil && conn != nil {
		err = c.options.ApplyTCP(conn)
		return network.NewTCPConn(conn), err
	}
	return nil, err
}
