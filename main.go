package main

import (
	"flag"
	"github.com/ghost0036/logT/lib/tcp"
	"sync"
	"time"
)

var serverName = flag.String("a", ":8080", "server Address")
var isServer = flag.Bool("server",true,"运行模式，true是服务器")
var wg sync.WaitGroup

func main() {
	flag.Parse()
	if *isServer  {
		go func() {
			wg.Add(1)
			tcpServer := tcp.TcpServer{}
			tcpServer.Open()
			tcpServer.Start()
		}()
		//等待一秒防止还没有执行到协程序
		time.Sleep(time.Second*1)
		wg.Wait()
	} else {
		tcpClient := tcp.TcpClient{ServerAdd: *serverName,}
		tcpClient.Open()
		tcpClient.Connection()

		sendDataTick := time.Tick(2 * time.Second)
		for {
			select {
			case <-sendDataTick:
				tcpClient.SendData("hello " + tcpClient.ServerAdd + ", now is " + time.Now().String())
			default:
				continue
			}
		}
	}
}
