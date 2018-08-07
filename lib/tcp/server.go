package tcp

import (
	"log"
	"net"
	"os"
)

var (
	server = ":8080"
)

func (tcpServer *TcpServer) Open() {
	//类似于初始化套接字，绑定端口
	hawkServer, err := net.ResolveTCPAddr("tcp", server)
	checkErr(err)
	tcpServer.hawkServer = hawkServer

	//侦听
	listen, err := net.ListenTCP("tcp", hawkServer)
	checkErr(err)
	tcpServer.listener = listen

	log.Println("Start server successful......listening on :", server)

}

func (tcpServer *TcpServer) Start() {
	defer tcpServer.listener.Close()

	log.Println("Tcp Server start!!!")
	for {
		conn, err := tcpServer.listener.AcceptTCP()
		log.Printf("Accept tcp from client: %s", conn.RemoteAddr().String())
		checkErr(err)
		// 每次建立一个连接就放到单独的协程内做处理
		SendData(conn,"you hava conntected!!")
		go ReceiverData(conn, func(rep *ReportPacket) {
			log.Println("Client===Server>", rep.Content)
		})
	}
}

func (tcpServer *TcpServer) Close() {
	tcpServer.listener.Close()
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}
