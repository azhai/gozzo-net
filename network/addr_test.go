package network

import "testing"

func TestLocalAddrs(t *testing.T) {
	addrs := GetLocalAddrs()
	for _, addr := range addrs {
		t.Log(addr.String())
	}
}

func TestLocalGroup(t *testing.T) {
	group := NewLocalAddrGroup()
	for i := 0; i < 20; i++ {
		addr := group.NextAddr()
		t.Log(addr.String())
	}
}
