package network

import (
	"io"
	"io/ioutil"
	"time"
)

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
