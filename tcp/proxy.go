package tcp

import (
	"io"

	"github.com/azhai/gozzo-net/network"
	"github.com/azhai/gozzo-net/udp"
)

type ProxyAction func(orig, relay *network.Conn)

// 端口转发代理.
type Proxy struct {
	options  network.TCPOptions
	dialplan *network.DialPlan
	server   *network.Server
}

// 创建代理
func NewProxy(dhost string, dport, sport uint16) *Proxy {
	addr, _ := network.NewTCPAddr(dhost, dport)
	return &Proxy{
		server:   network.NewServer("", sport, 0),
		dialplan: network.NewDialPlan(addr, nil, 10),
		options:  network.DefaultTCPOptions,
	}
}

func (p *Proxy) Run(kind string, action ProxyAction) (err error) {
	events := network.Events{}
	events.Process = func(s *network.Server, c *network.Conn) {
		defer s.CloseConn(c, nil)
		// 创建TCP客户端，连接到真正的TCP/UDP服务器和端口
		client := NewClient(p.dialplan, p.options)
		network.Reconnect(client, true, 3)
		if conn := client.GetConn(); conn != nil {
			action(c, conn)
		}
	}
	if kind == "udp" {
		err = udp.NewServer(p.server).Run(events)
	} else {
		err = NewServer(p.server).Run(events)
	}
	return
}

func RelayData(orig, relay *network.Conn) {
	defer relay.Close()
	go io.Copy(relay.GetRawConn(), orig.GetReader()) // 复制上报数据
	io.Copy(orig.GetRawConn(), relay.GetReader())    // 复制服务端回应
}
