package unix

import (
	"io"
	"net"

	"github.com/azhai/gozzo-net/network"
	"github.com/azhai/gozzo-net/tcp"
	"github.com/azhai/gozzo-net/udp"
)

type IRouter interface {
	Dispatch(c *network.Conn) (string, *network.DialPlan)
}

type Relayer struct {
	Kind string
	*network.DialPlan
}

func NewRelayer(addr net.Addr) *Relayer {
	dp := network.NewDialPlan(addr, nil, 10)
	return &Relayer{Kind: "tcp", DialPlan: dp}
}

func NewUnixRelayer(addr net.Addr) *Relayer {
	r := NewRelayer(addr)
	r.Kind = "unix"
	return r
}

func (r *Relayer) Dispatch(c *network.Conn) (string, *network.DialPlan) {
	return r.Kind, r.DialPlan
}

type ProxyAction func(s *network.Server, orig, relay *network.Conn)

// 原样复制输入和输出
func RelayData(s *network.Server, orig, relay *network.Conn) {
	defer relay.Close()
	go io.Copy(orig.GetRawConn(), relay.GetReader()) // 复制服务端回应
	// NOTICE: 与上面一行不能对调，否则无法知道客户端关闭了
	io.Copy(relay.GetRawConn(), orig.GetReader()) // 复制上报数据
}

// 转发代理
type Proxy struct {
	kind    string
	Options network.TCPOptions
	*network.Server
}

// 创建TCP/UDP代理
func NewProxy(kind, host string, port uint16) *Proxy {
	opts := network.DefaultTCPOptions
	serv := network.NewPortServer(host, port)
	return &Proxy{kind: kind, Options: opts, Server: serv}
}

func (p *Proxy) CreateClient(kind string, dp *network.DialPlan) (client network.IClient) {
	if dp == nil {
		return
	}
	if kind == "tcp" {
		client = tcp.NewClient(dp, p.Options)
	} else if kind == "udp" {
		client = udp.NewClient(dp, p.Options.Options)
	} else { // unix
		client = NewClient(dp, p.Options.Options)
	}
	return
}

func (p *Proxy) CreateProcess(router IRouter, action ProxyAction) network.ProcessFunc {
	return func(s *network.Server, c *network.Conn) {
		// 创建客户端，连接到真正的服务器
		client := p.CreateClient(router.Dispatch(c))
		if client == nil {
			return
		}
		defer client.Close()
		network.Reconnect(client, true, 3)
		if conn := client.GetConn(); conn != nil {
			action(s, c, conn)
		}
	}
}

func (p *Proxy) Run(events network.Events) (err error) {
	if p.kind == "udp" {
		err = udp.NewServer(p.Server).Run(events)
	} else {
		err = tcp.NewServer(p.Server).Run(events)
	}
	return
}
