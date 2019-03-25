package udp

import (
	"github.com/azhai/gozzo-net/network"
)

// UDP 客户端
type UDPClient struct {
	options  network.Options
	dialplan *network.DialPlan
	Conn     *network.Conn
}

// 创建UDP客户端
func NewClient(plan *network.DialPlan, opts network.Options) *UDPClient {
	return &UDPClient{dialplan: plan, options: opts}
}

func (c *UDPClient) Close() error {
	if c.Conn == nil {
		return nil
	}
	return c.Conn.Close()
}

func (c *UDPClient) GetConn() *network.Conn {
	return c.Conn
}

func (c *UDPClient) SetConn(conn *network.Conn) {
	c.Conn = conn
}

func (c *UDPClient) Dialing() (*network.Conn, error) {
	conn, err := c.dialplan.DialUDP()
	if err == nil && conn != nil {
		err = c.options.ApplyConn(conn)
		return network.NewUDPConn(conn), err
	}
	return nil, err
}
