package unix

import (
	"os"

	"github.com/azhai/gozzo-net/network"
	"github.com/azhai/gozzo-utils/common"
)

// Unix socket 客户端.
type UnixClient struct {
	options  network.Options
	dialplan *network.DialPlan
	Conn     *network.Conn
}

// 创建Unix客户端
func NewClient(plan *network.DialPlan, opts network.Options) *UnixClient {
	return &UnixClient{dialplan: plan, options: opts}
}

func (c *UnixClient) Close() error {
	var err error
	if c.Conn != nil {
		err = c.Conn.Close()
	}
	filename := c.dialplan.LocalAddr.String()
	if _, exists := common.FileSize(filename); exists {
		err = os.Remove(filename)
	}
	return err
}

func (c *UnixClient) GetPlan() *network.DialPlan {
	return c.dialplan
}

func (c *UnixClient) GetConn() *network.Conn {
	return c.Conn
}

func (c *UnixClient) SetConn(conn *network.Conn) {
	c.Conn = conn
}

func (c *UnixClient) Dialing() (*network.Conn, error) {
	conn, err := c.dialplan.DialUnix()
	if err == nil && conn != nil {
		err = c.options.ApplyConn(conn)
		return network.NewUnixConn(conn), err
	}
	return nil, err
}
