package network

import (
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/azhai/gozzo-utils/metrics"
)

// 获取所有局域网IP，除了127.0.0.1，未排序
func GetLocalAddrs() (result []*net.IPNet) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return result
	}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok {
			if ip.IP.IsLoopback() || ip.IP.To4() == nil {
				continue
			}
			result = append(result, ip)
		}
	}
	return result
}

// 局域网IP的循环列表
type LocalAddrRing struct {
	ring       *metrics.Ring
	LocalAddrs []*net.IPNet
	TCPAddrs   []*net.TCPAddr
}

func NewLocalAddrRing() *LocalAddrRing {
	addrs := GetLocalAddrs()
	ring := metrics.NewRing(len(addrs))
	return &LocalAddrRing{ring: ring, LocalAddrs: addrs}
}

func (r *LocalAddrRing) Translate(i, stop int) {
	for i <= stop {
		addr := &net.TCPAddr{IP: r.LocalAddrs[i].IP}
		r.TCPAddrs = append(r.TCPAddrs, addr)
		i++
	}
}

func (r *LocalAddrRing) NextAddr() (addr *net.IPNet) {
	if curr := r.ring.Next(); curr >= 0 {
		addr = r.LocalAddrs[curr]
	}
	return
}

func (r *LocalAddrRing) NextTCPAddr() (addr *net.TCPAddr) {
	if curr := r.ring.Next(); curr >= 0 {
		if i := len(r.TCPAddrs); i <= curr {
			r.Translate(i, curr)
		}
		addr = r.TCPAddrs[curr]
	}
	return
}

func (r *LocalAddrRing) GetTCPAddrs() []*net.TCPAddr {
	if count := len(r.LocalAddrs); count > 0 {
		r.Translate(len(r.TCPAddrs), count - 1)
	}
	return r.TCPAddrs
}

type IClient interface {
	GetConn() *Conn
	SetConn(conn *Conn)
	Dialing() (*Conn, error)
	Close() error
}

// 按需重连服务端
func Reconnect(client IClient, force bool, retries int) (n int, err error) {
	conn := client.GetConn()
	if conn != nil && conn.IsActive {
		if force {
			conn.Close()
		} else {
			return
		}
	}
	for n < retries { //重试几次
		n++
		if conn, err = client.Dialing(); err == nil {
			client.SetConn(conn)
			return
		}
		time.Sleep(time.Duration(n) * time.Second)
	}
	return
}

func SendData(client IClient, data []byte) (err error) {
	if _, err = Reconnect(client, false, 3); err == nil {
		if conn := client.GetConn(); conn != nil {
			err = conn.QuickSend(data)
		}
	}
	return
}

func Discard(client IClient) (err error) {
	if conn := client.GetConn(); conn != nil {
		if reader := conn.GetReader(); reader != nil {
			_, err = io.Copy(ioutil.Discard, reader)
		}
	}
	return
}
