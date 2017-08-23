package mysql

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

/*
mysql 相关数据包信息
*/
var DEFAULT_CAPABILITY uint32 = CLIENT_LONG_PASSWORD | CLIENT_LONG_FLAG | CLIENT_CONNECT_WITH_DB | CLIENT_PROTOCOL_41 | CLIENT_TRANSACTIONS | CLIENT_SECURE_CONNECTIO

var baseConnId uint32 = 10000

/*
client 连接信息
*/
type ClientConn struct {
	sync.Mutex

	//数据包操作指针
	pkg *PacketIO

	//连接对象
	c          net.Conn
	capability uint32

	connectionId uint32

	status       uint16
	collation    CollationId
	charset      string
	user         string
	db           string
	salt         []byte
	closed       bool
	lastInsertId int64
	affectedRows int64
	stmtId       uint32
}

/*
初始化客户端连接信息
*/
func NewClientConn(co net.Conn) *ClientConn {
	c := new(ClientConn)
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

//server 发送初始化握手包
func (c *ClientConn) writeInitialHandshake() error {

	fmt.Println("send initial handshake packet")
	data := make([]byte, 4, 128)
	//协议版本号 version 10
	data = append(data, ProtocolVersion)

	//server version[00]
	data = append(data, ServerVersion...)
	data = append(data, 0)

	//connection id
	data = append(data, byte(c.connectionId), byte(c.connectionId>>8), byte(c.connectionId>>16), byte(c.connectionId>>24))

	//auth-plugin-data-part-1
	data = append(data, c.salt[0:8]...)
	//filter [00]
	data = append(data, 0)

	//capability flag lower 2 bytes, using default capability here
	data = append(data, byte(DEFAULT_CAPABILITY), byte(DEFAULT_CAPABILITY>>8))

	//charset, utf-8 default
	data = append(data, uint8(DEFAULT_COLLATION_ID))

	//status
	data = append(data, byte(c.status), byte(c.status>>8))
	//below 13 byte may not be used
	//capability flag upper 2 bytes, using default capability here
	data = append(data, byte(DEFAULT_CAPABILITY>>16), byte(DEFAULT_CAPABILITY>>24))

	//filter [0x15], for wireshark dump, value is 0x15
	data = append(data, 0x15)

	//reserved 10 [00]
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)

	//auth-plugin-data-part-2
	data = append(data, c.salt[8:]...)

	//filter [00]
	data = append(data, 0)

	return c.pkg.WritePacket(data)
}

//server 解析初始化握手包的反馈信息
func (c *ClientConn) readHandshakeResponse() error {

	return nil
}
