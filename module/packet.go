package module

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

const (
	UP_ACTION             = 1 //上传
	DOWN_ACTION           = 2 //下载
	OTHER_ACTION          = 4 //其他
	defaultReaderSize     = 8 * 1024
	MaxPayloadLen     int = 1<<24 - 1
)

/*
数据包格式
*/
// type Packet struct {
// 	Header      uint32  //包头
// 	HeaderLengh int32  //包长
// 	Data        []byte //body
// }

type PacketIO struct {
	rb       *bufio.Reader
	wb       io.Writer
	Sequence uint8 //交互包序列号
}

func NewPacket(conn net.Conn) *PacketIO {
	p := new(PacketIO)
	p.rb = bufio.NewReaderSize(conn, defaultReaderSize)
	p.wb = conn
	p.Sequence = 0
	return p
}

/*
拆包
*/
func (p *PacketIO) ReadPacket() ([]byte, error) {
	//数据包头 4个字节
	header := []byte{0, 0, 0, 0}
	//读取4个字节
	if _, err := io.ReadFull(p.rb, header); err != nil {
		return nil, err
	}

	//获取前3个字节长度表示的包长
	length := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)
	if length < 1 {
		return nil, fmt.Errorf("invalid payload length:%d", length)
	}

	//获取第4个字节表示的包序列号
	sequence := uint8(header[3])
	if sequence != p.Sequence {
		return nil, fmt.Errorf("invalid packet sequnece %d !=%d", sequence, p.Sequence)
	}

	//response client packet sequence +1
	p.Sequence++

	//创建body缓冲区,并读取payload
	data := make([]byte, length)

	if _, err := io.ReadFull(p.rb, data); err != nil {
		return nil, fmt.Errorf("bad connection! error:%s", err)
	}

	//处理粘包
	//要读取的数据小于maxpayloadlen
	if length < MaxPayloadLen {
		return data, nil
	}
	//要读取的数据大于maxpayloadlen,再次递归读取
	var buf []byte

	buf, err := p.ReadPacket()
	if err != nil {
		return nil, fmt.Errorf("bad connection!")
	} else {

		return append(data, buf...), nil
	}

	return data, nil
}
