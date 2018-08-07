package tcp

import (
	"net"
	"time"
	"encoding/json"
	"log"
	"math/rand"
	"bufio"
	"io"
	"hash/crc32"
)

//数据包类型
const (
	HEART_BEAT_PACKET = 0x00
	REPORT_PACKET     = 0x01
)

//数据包
type Packet struct {
	PacketType    byte
	PacketContent []byte
}

//心跳包
type HeartPacket struct {
	Version   string `json:"version"`
	Timestamp int64  `json:"timestamp"`
}

//数据包
type ReportPacket struct {
	Content   string `json:"content"`
	Rand      int    `json:"rand"`
	Timestamp int64  `json:"timestamp"`
}

//客户端对象
type TcpClient struct {
	ServerAdd  string
	Conn *net.TCPConn
	hawkServer *net.TCPAddr
	stopChan   chan struct{}
}

//服务端对象
type TcpServer struct {
	listener   *net.TCPListener
	hawkServer *net.TCPAddr
}

type clientCommand struct {
	commond string
	options []string
	function  func(args... interface{})
}
//封包
func SendData(connection *net.TCPConn,data string) {
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
	connection.Write(EnPackSendData(sendBytes))
	log.Println("Send data success!")
}

//解包
func ReceiverData(conn net.Conn,f func(rep *ReportPacket)) {
	// close connection before exit
	defer conn.Close()
	// 状态机状态
	state := 0x00
	// 数据包长度
	length := uint16(0)
	// crc校验和
	crc16 := uint16(0)
	var recvBuffer []byte
	// 游标
	cursor := uint16(0)
	bufferReader := bufio.NewReader(conn)
	//状态机处理数据
	for {
		recvByte, err := bufferReader.ReadByte()
		if err != nil {
			//这里因为做了心跳，所以就没有加deadline时间，如果客户端断开连接
			//这里ReadByte方法返回一个io.EOF的错误，具体可考虑文档
			if err == io.EOF {
				log.Printf("conn %s is close!\n", conn.RemoteAddr().String())
			}
			//在这里直接退出goroutine，关闭由defer操作完成
			return
		}
		//进入状态机，根据不同的状态来处理
		switch state {
		case 0x00:
			if recvByte == 0xFF {
				state = 0x01
				//初始化状态机
				recvBuffer = nil
				length = 0
				crc16 = 0
			} else {
				state = 0x00
			}
			break
		case 0x01:
			if recvByte == 0xFF {
				state = 0x02
			} else {
				state = 0x00
			}
			break
		case 0x02:
			length += uint16(recvByte) * 256
			state = 0x03
			break
		case 0x03:
			length += uint16(recvByte)
			// 一次申请缓存，初始化游标，准备读数据
			recvBuffer = make([]byte, length)
			cursor = 0
			state = 0x04
			break
		case 0x04:
			//不断地在这个状态下读数据，直到满足长度为止
			recvBuffer[cursor] = recvByte
			cursor++
			if cursor == length {
				state = 0x05
			}
			break
		case 0x05:
			crc16 += uint16(recvByte) * 256
			state = 0x06
			break
		case 0x06:
			crc16 += uint16(recvByte)
			state = 0x07
			break
		case 0x07:
			if recvByte == 0xFF {
				state = 0x08
			} else {
				state = 0x00
			}
		case 0x08:
			if recvByte == 0xFE {
				//执行数据包校验
				if (crc32.ChecksumIEEE(recvBuffer)>>16)&0xFFFF == uint32(crc16) {
					var packet Packet
					//把拿到的数据反序列化出来
					json.Unmarshal(recvBuffer, &packet)
					//新开协程处理数据
					go processRecvData(&packet, conn,f)
				} else {
					log.Println("数据校验失败，丢弃数据!")
				}
			}
			//状态机归位,接收下一个包
			state = 0x00
		}
	}
}

func processRecvData(packet *Packet, conn net.Conn,f func(repo *ReportPacket)) {
	switch packet.PacketType {
	case HEART_BEAT_PACKET:
		var beatPacket HeartPacket
		json.Unmarshal(packet.PacketContent, &beatPacket)
		log.Printf("recieve heat beat from [%s] ,data is [%v]\n", conn.RemoteAddr().String(), beatPacket)
		return
	case REPORT_PACKET:
		var reportPacket ReportPacket
		json.Unmarshal(packet.PacketContent, &reportPacket)
		//log.Printf("recieve report data from [%s] ,data is [%v]\n", conn.RemoteAddr().String(), reportPacket)
		f(&reportPacket)
		return
	}
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
