package network

import "testing"

func TestLocalAddrs(t *testing.T) {
	addrs := GetLocalAddrs()
	for _, addr := range addrs {
		t.Log(addr.String())
	}
}

func TestLocalRing(t *testing.T) {
	group := NewLocalAddrRing()
	for i := 0; i < 10; i++ {
		addr1 := group.NextAddr()
		t.Log(addr1.Network(), addr1.String())
		addr2 := group.NextAddr()
		t.Log(addr2.Network(), addr2.String())
		taddr := group.NextTCPAddr()
		t.Log(taddr.Network(), taddr.String())
	}
}
