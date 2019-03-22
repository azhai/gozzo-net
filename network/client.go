package network

import (
	"io"
	"io/ioutil"
	"net"
	"sync"
	"time"
)

// 获取所有局域网IP，除了127.0.0.1，未排序
func GetLocalAddrs() []net.Addr {
	var result []net.Addr
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return result
	}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok {
			if ip.IP.IsLoopback() || ip.IP.To4() == nil {
				continue
			}
			result = append(result, addr)
		}
	}
	return result
}

// 局域网IP的循环列表
type LocalAddrGroup struct {
	pointer    int
	mutex      *sync.RWMutex
	LocalAddrs []net.Addr
}

func NewLocalAddrGroup() *LocalAddrGroup {
	return &LocalAddrGroup{
		mutex:      new(sync.RWMutex),
		LocalAddrs: GetLocalAddrs(),
	}
}

func (g *LocalAddrGroup) NextAddr() (addr net.Addr) {
	var size int
	if size = len(g.LocalAddrs); size == 0 {
		return
	}
	g.mutex.RLock()
	if g.pointer >= size {
		g.pointer = 0
	}
	addr = g.LocalAddrs[g.pointer]
	g.pointer++
	g.mutex.RUnlock()
	return
}

type IClient interface {
	GetConn() *Conn
	SetConn(conn *Conn)
	Dialing() (*Conn, error)
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
