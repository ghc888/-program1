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
func (s *GateServer) NewClientConn(co net.Conn) *ClientConn {

	c := new(ClientConn)

	tcpConn := co.(*net.TCPConn)
	tcpConn.SetNoDelay(false)
	tcpConn.SetKeepAlive(true)
	c.c = tcpConn
	//初始化包序列号
	c.pkg.Sequence = 0

	//初始化连接id  自增id
	c.connectionId = atomic.AddUint32(&baseConnId, 1)
	c.status = mysql.SERVER_STATUS_AUTOCOMMIT
	c.salt, _ = mysql.RandomBuf(20)
	c.closed = false
	c.charset = mysql.DEFAULT_CHARSET
	c.collation = mysql.DEFAULT_COLLATION_ID
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
	err = newconn.Handshake()
	if err != nil {
		fmt.Println("shandshake connection err:", err)
	}

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
