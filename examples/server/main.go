package main

import (
	"fmt"
	"os"
	"time"

	"github.com/azhai/gozzo-net/network"
	"github.com/azhai/gozzo-net/tcp"
	"github.com/azhai/gozzo-pck/match"
	"github.com/azhai/gozzo-utils/common"
	"github.com/azhai/gozzo-utils/metrics"
	"github.com/azhai/gozzo-utils/queue"
)

func run() {
	events := network.Events{}

	// 建立新连接
	events.Opened = func(s *network.Server, c *network.Conn) error {
		reporter.IncrCount("opened", 1)
		if buff := conf.Server.GetBuffSize(); buff > 0 {
			return c.GetRawConn().SetReadBuffer(buff)
		}
		return nil
	}

	// 接收数据前
	events.Prepare = func(c *network.Conn, input chan<- []byte) error {
		sc := match.NewFixedSplitCreator(conf.Proto.GetBuffSize())
		sp := match.NewSplitMatcher(sc.GetSplit())
		return sp.SplitStream(c.GetReader(), input) // 分割，处理粘包
	}

	// 设备上报数据
	events.Receive = func(c *network.Conn, data []byte, saved bool) (imei string) {
		counter := reporter.IncrCount("received", 1)
		if logger != nil {
			addr := c.GetRawConn().RemoteAddr().String()
			logger.Debug(addr, "\t", common.Bin2Hex(data))
		}
		msg := queue.NewMessage(data)
		err := conf.Rabbit.Push(int(counter), msg)
		CheckError(err)
		return
	}

	// 连接断开，包括设备主动断开，和通讯异常被动断开（每个连接只执行一次）
	events.Closed = func(s *network.Server, c *network.Conn, err error) {
		// 记录断开数，数字可能比实际的少
		reporter.IncrCount("closed", 1)
		CheckError(err)
	}

	serv := network.NewPortServer(conf.Server.Host, uint16(conf.Server.Port))
	if tick := conf.Server.Tick; tick > 0 {
		// 定时任务，输出统计数据
		serv.SetTickInterval(tick)
		events.Tick = func(t time.Time) {
			if logger != nil {
				logger.Info(metrics.StatSnap(reporter, false))
			}
			fmt.Print("\r", metrics.StatSnap(reporter, true))
			_ = os.Stdout.Sync()
		}
	}

	// 运行服务器监听端口
	err := tcp.NewServer(serv).Run(events)
	CheckError(err)
}
