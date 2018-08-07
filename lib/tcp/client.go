package tcp

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"time"
)

func (tcpClient *TcpClient) Open() {

	//拿到服务器地址信息
	hawkServer, err := net.ResolveTCPAddr("tcp", tcpClient.ServerAdd)
	if err != nil {
		log.Printf("Server [%s] resolve error: [%s]", server, err.Error())
		os.Exit(1)
	}
	tcpClient.hawkServer = hawkServer

	tcpClient.stopChan = make(chan struct{})
}

//建立连接
func (tcpClient *TcpClient) Connection() {

	connection, err := net.DialTCP("tcp", nil, tcpClient.hawkServer)
	if err != nil {
		log.Fatal("Connect to server error: [%s]", err.Error())
		os.Exit(1)
	}
	tcpClient.Conn = connection

	go  ReceiverData(tcpClient.Conn, func(rep *ReportPacket) {
		log.Println("Server=====>client:",rep.Content)
	})

	//发送心跳的goroutine
	go func() {
		heartBeatTick := time.Tick(10 * time.Second)
		for {
			select {
			case <-heartBeatTick:
				tcpClient.sendHeartPacket()
			case <-tcpClient.stopChan:
				return
			}
		}
	}()
}


//发送心跳包，与发送数据包一样
func (client *TcpClient) sendHeartPacket() {
	heartPacket := HeartPacket{
		Version:   "1.0",
		Timestamp: time.Now().Unix(),
	}
	packetBytes, err := json.Marshal(heartPacket)
	if err != nil {
		log.Println(err.Error())
	}
	packet := Packet{
		PacketType:    HEART_BEAT_PACKET,
		PacketContent: packetBytes,
	}
	sendBytes, err := json.Marshal(packet)
	if err != nil {
		log.Println(err.Error())
	}
	client.Conn.Write(EnPackSendData(sendBytes))
	log.Println("Send heartbeat data success!")
}
