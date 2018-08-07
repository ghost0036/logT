package main

import (
	"flag"
	"github.com/ghost0036/logT/lib/tcp"
	"sync"
	"time"
)

var serverName = flag.String("-server", "0.0.0.0:8080", "server Address")
var wg sync.WaitGroup

func main() {
	if serverName != nil {
		tcpClient := tcp.TcpClient{ServerAdd: "127.0.0.1:8080"}
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
	} else {
		go func() {
			wg.Add(1)
			tcpServer := tcp.TcpServer{}
			tcpServer.Open()
			tcpServer.Start()
		}()

		wg.Wait()
	}
}
