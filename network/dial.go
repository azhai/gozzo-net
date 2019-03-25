package network

import (
	"fmt"
	"net"
	"time"
)

// 根据IP和端口创建TCP地址
func NewTCPAddr(host string, port uint16) *net.TCPAddr {
	address := fmt.Sprintf("%s:%d", host, port)
	if addr, err := net.ResolveTCPAddr("tcp", address); err == nil {
		return addr
	}
	return nil
}

// 将网络地址转为TCP地址
func GetTCPAddr(addr net.Addr) (tcpAddr *net.TCPAddr) {
	if addr.Network() == "tcp" {
		tcpAddr = addr.(*net.TCPAddr)
	} else {
		tcpAddr, _ = net.ResolveTCPAddr("tcp", addr.String())
	}
	return
}

// 将网络地址转为UDP地址
func GetUDPAddr(addr net.Addr) (udpAddr *net.UDPAddr) {
	if addr.Network() == "udp" {
		udpAddr = addr.(*net.UDPAddr)
	} else {
		udpAddr, _ = net.ResolveUDPAddr("udp", addr.String())
	}
	return
}

// 将文件名转为Unix socket地址
func NewUnixAddr(filename string) (addr *net.UnixAddr) {
	addr, err := net.ResolveUnixAddr("unix", filename)
	if err != nil {
		panic(err)
	}
	return
}

// 拨号计划
type DialPlan struct {
	Timeout    time.Duration
	LocalAddr  net.Addr
	RemoteAddr net.Addr
}

func NewDialPlan(remote, local net.Addr, timeout int) *DialPlan {
	return &DialPlan{
		Timeout:    time.Duration(timeout) * time.Second,
		LocalAddr:  local,
		RemoteAddr: remote,
	}
}

// 拨号得到网络连接
func (dp *DialPlan) Dial(kind string) (net.Conn, error) {
	address := dp.RemoteAddr.String()
	if dp.LocalAddr == nil {
		return net.DialTimeout(kind, address, dp.Timeout)
	} else {
		d := &net.Dialer{LocalAddr: dp.LocalAddr, Timeout: dp.Timeout}
		return d.Dial(kind, address)
	}
}

// 拨号得到TCP连接
func (dp *DialPlan) DialTCP() (*net.TCPConn, error) {
	if dp.Timeout <= 0 {
		var laddr *net.TCPAddr
		if dp.LocalAddr != nil {
			laddr = GetTCPAddr(dp.LocalAddr)
		}
		addr := GetTCPAddr(dp.RemoteAddr)
		return net.DialTCP("tcp", laddr, addr)
	}
	conn, err := dp.Dial("tcp")
	if err == nil {
		return conn.(*net.TCPConn), err
	}
	return nil, err
}

// 拨号得到UDP连接
func (dp *DialPlan) DialUDP() (*net.UDPConn, error) {
	if dp.Timeout <= 0 {
		var laddr *net.UDPAddr
		if dp.LocalAddr != nil {
			laddr = GetUDPAddr(dp.LocalAddr)
		}
		addr := GetUDPAddr(dp.RemoteAddr)
		return net.DialUDP("udp", laddr, addr)
	}
	conn, err := dp.Dial("udp")
	if err == nil {
		return conn.(*net.UDPConn), err
	}
	return nil, err
}

// 拨号得到Unix连接
func (dp *DialPlan) DialUnix() (*net.UnixConn, error) {
	if dp.Timeout <= 0 {
		laddr := dp.LocalAddr.(*net.UnixAddr)
		addr := dp.RemoteAddr.(*net.UnixAddr)
		return net.DialUnix("unix", laddr, addr)
	}
	conn, err := dp.Dial("unix")
	if err == nil {
		return conn.(*net.UnixConn), err
	}
	return nil, err
}

func (dp *DialPlan) SetPortRemote(host string, port uint16) {
	dp.RemoteAddr = NewTCPAddr(host, port)
}

func (dp *DialPlan) SetUnixRemote(filename string) {
	dp.RemoteAddr = NewUnixAddr(filename)
}
