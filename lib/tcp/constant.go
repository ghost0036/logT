package tcp

import "net"

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
	connection *net.TCPConn
	hawkServer *net.TCPAddr
	stopChan   chan struct{}
}

//服务端对象
type TcpServer struct {
	listener   *net.TCPListener
	hawkServer *net.TCPAddr
}
