package server

import (
	"fmt"
	"net"
	"program1/mysql"
	"sync/atomic"
)

/*
入口server
*/
type GateServer struct {
	addr     string //侦听地址
	listener net.Listener
	running  bool //运行状态
}

func NewServer() (*GateServer, error) {
	s := new(GateServer)
	var err error
	netProto := "tcp"
	s.addr = ":8080"
	s.listener, err = net.Listen(netProto, s.addr)
	if err != nil {
		return nil, err
	}
	return s, nil
}

/*
初始化客户端连接信息
*/
func (s *GateServer) NewClientConn(co net.Conn) *mysql.ClientConn {

	c := new(mysql.ClientConn)
	tcpConn := co.(*net.TCPConn)
	tcpConn.SetNoDelay(false)
	tcpConn.SetKeepAlive(true)
	c.pkg = NewPacketIO(tcpConn)

	//初始化包序列号
	c.pkg.Sequence = 0

	//初始化连接id  自增id
	c.connectionId = atomic.AddUint32(&baseConnId, 1)
	c.status = SERVER_STATUS_AUTOCOMMIT
	c.salt, _ = RandomBuf(20)
	c.closed = false
	c.charset = DEFAULT_CHARSET
	c.collation = DEFAULT_COLLATION_ID
	c.stmtId = 0
	return c
}

func (s *GateServer) handleConnectionV2(con net.Conn) {
	defer con.Close()

	clientHost, _, err := net.SplitHostPort(con.RemoteAddr().String())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("client ip allow conenction server", clientHost)

	newconn := s.NewClientConn(con)

	data, err := newconn.ReadPacket()
	if err != nil {
		fmt.Println("read client message error:", err)
	}
	pos := 0
	fmt.Println("message topic:", data[pos])
	pos++
	fmt.Println("message :", string(data[pos:]))
}

func (s *GateServer) GRun() {

	for {

		con, err := s.listener.Accept()
		if err != nil {
			fmt.Println("server accept error:", err)
			continue
		}
		go s.handleConnectionV2(con)
	}
}
