package tcp

import (
	"bufio"
	"encoding/json"
	"hash/crc32"
	"log"
	"math/rand"
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
	tcpClient.connection = connection

	go tcpClient.receivePackets()

	//发送心跳的goroutine
	go func() {
		heartBeatTick := time.Tick(2 * time.Second)
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

// 接收数据包
func (tcpClient *TcpClient) receivePackets() {
	reader := bufio.NewReader(tcpClient.connection)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			//在这里也请处理如果服务器关闭时的异常
			close(tcpClient.stopChan)
			break
		}
		log.Println("从服务端接收到数据：" + msg)
	}
}

//发送数据包
//仔细看代码其实这里做了两次json的序列化，有一次其实是不需要的
func (client *TcpClient) SendData(data string) {
	reportPacket := ReportPacket{
		Content:   data,
		Timestamp: time.Now().Unix(),
		Rand:      rand.Int(),
	}
	packetBytes, err := json.Marshal(reportPacket)
	if err != nil {
		log.Println(err.Error())
	}

	//这一次其实可以不需要，在封包的地方把类型和数据传进去即可
	packet := Packet{
		PacketType:    REPORT_PACKET,
		PacketContent: packetBytes,
	}
	sendBytes, err := json.Marshal(packet)
	if err != nil {
		log.Println(err.Error())
	}
	//发送
	client.connection.Write(EnPackSendData(sendBytes))
	log.Println("Send data success!")
}

//使用的协议与服务器端保持一致
func EnPackSendData(sendBytes []byte) []byte {
	packetLength := len(sendBytes) + 8
	result := make([]byte, packetLength)
	result[0] = 0xFF
	result[1] = 0xFF
	result[2] = byte(uint16(len(sendBytes)) >> 8)
	result[3] = byte(uint16(len(sendBytes)) & 0xFF)
	copy(result[4:], sendBytes)
	sendCrc := crc32.ChecksumIEEE(sendBytes)
	result[packetLength-4] = byte(sendCrc >> 24)
	result[packetLength-3] = byte(sendCrc >> 16 & 0xFF)
	result[packetLength-2] = 0xFF
	result[packetLength-1] = 0xFE
	return result
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
	client.connection.Write(EnPackSendData(sendBytes))
	log.Println("Send heartbeat data success!")
}
