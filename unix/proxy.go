package unix

import (
	"io"
	"net"

	"github.com/azhai/gozzo-net/network"
	"github.com/azhai/gozzo-net/tcp"
	"github.com/azhai/gozzo-net/udp"
)

type IRouter interface {
	Dispatch(c *network.Conn) *network.DialPlan
}

type Relayer struct {
	*network.DialPlan
}

func NewRelayer(addr net.Addr) *Relayer {
	return &Relayer{
		DialPlan: network.NewDialPlan(addr, nil, 10),
	}
}

func (r *Relayer) Dispatch(c *network.Conn) *network.DialPlan {
	return r.DialPlan
}

type ProxyAction func(s *network.Server, orig, relay *network.Conn)

// 原样复制输入和输出
func RelayData(s *network.Server, orig, relay *network.Conn) {
	defer relay.Close()
	go io.Copy(relay.GetRawConn(), orig.GetReader()) // 复制上报数据
	io.Copy(orig.GetRawConn(), relay.GetReader())    // 复制服务端回应
}

// 转发代理
type Proxy struct {
	kind       string
	Options    network.TCPOptions
	*network.Server
}

// 创建代理
func NewProxy(kind, host string, port uint16) *Proxy {
	var serv *network.Server
	opts := network.DefaultTCPOptions
	if kind == "tcp" {
		serv = network.NewPortServer(host, port)
	} else {
		serv = network.NewUnixServer(host)
	}
	return &Proxy{kind: kind, Options: opts, Server: serv}
}

func (p *Proxy) CreateClient(dp *network.DialPlan) (client network.IClient) {
	if p.kind == "tcp" {
		client = tcp.NewClient(dp, p.Options)
	} else if p.kind == "udp" {
		client = udp.NewClient(dp, p.Options.Options)
	} else {
		client = NewClient(dp, p.Options.Options)
	}
	return
}

func (p *Proxy) CreateProcess(router IRouter, action ProxyAction) network.ProcessFunc {
	return func(s *network.Server, c *network.Conn) {
		var dp *network.DialPlan
		if dp = router.Dispatch(c); dp == nil {
			return
		}
		// 创建客户端，连接到真正的服务器
		client := p.CreateClient(dp)
		defer client.Close()
		network.Reconnect(client, true, 3)
		if conn := client.GetConn(); conn != nil {
			action(s, c, conn)
		}
	}
}

func (p *Proxy) Run(kind string, events network.Events) (err error) {
	if kind == "udp" {
		err = udp.NewServer(p.Server).Run(events)
	} else {
		err = tcp.NewServer(p.Server).Run(events)
	}
	return
}
