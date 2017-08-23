package mysql

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
)

// const (
// 	defaultReaderSize     = 8 * 1024
// 	MaxPayloadLen     int = 1<<24 - 1
// )

/*

数据包的 IO信息
Conn is the base class to handle MySQL protocol.
*/
type PacketIO struct {
	br       *bufio.Reader //read buffer
	wb       io.Writer     //writer
	Sequence uint8
}

/*
初始化连接信息
*/
func NewPacketIO(conn net.Conn) *PacketIO {
	p := new(PacketIO)

	p.br = bufio.NewReaderSize(conn, 4096)
	p.wb = conn
	return p
}

/*
拆包
*/
func (p *PacketIO) ReadPacket() ([]byte, error) {
	var buf bytes.Buffer

	if err := p.ReadPacketTo(&buf); err != nil {
		return nil, err
	} else {
		return buf.Bytes(), nil
	}
}

/*
read to some where
*/
func (p *PacketIO) ReadPacketTo(w io.Writer) error {
	header := []byte{0, 0, 0, 0}

	if _, err := io.ReadFull(p.br, header); err != nil {
		return err
	}

	length := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)
	if length < 1 {
		return fmt.Errorf("invalid payload length %d", length)
	}

	sequence := uint8(header[3])

	if sequence != p.Sequence {
		return fmt.Errorf("invalid sequence %d != %d", sequence, p.Sequence)
	}

	p.Sequence++

	if n, err := io.CopyN(w, p.br, int64(length)); err != nil {
		return errors.New("connection was bad")
	} else if n != int64(length) {
		return errors.New("connection was bad")
	} else {
		if length < MaxPayloadLen {
			return nil
		}

		if err := p.ReadPacketTo(w); err != nil {
			return err
		}
	}

	return nil
}

/*
封包
*/
// data already has 4 bytes header
// will modify data inplace
func (p *PacketIO) WritePacket(data []byte) error {
	length := len(data) - 4

	for length >= MaxPayloadLen {
		data[0] = 0xff
		data[1] = 0xff
		data[2] = 0xff

		data[3] = p.Sequence

		if n, err := p.wb.Write(data[:4+MaxPayloadLen]); err != nil {
			return errors.New("connection was bad")
		} else if n != (4 + MaxPayloadLen) {
			return errors.New("connection was bad")
		} else {
			p.Sequence++
			length -= MaxPayloadLen
			data = data[MaxPayloadLen:]
		}
	}

	data[0] = byte(length)
	data[1] = byte(length >> 8)
	data[2] = byte(length >> 16)
	data[3] = p.Sequence

	if n, err := p.wb.Write(data); err != nil {
		return errors.New("connection was bad")
	} else if n != len(data) {
		return errors.New("connection was bad")
	} else {
		p.Sequence++
		return nil
	}
}

func (p *PacketIO) WritePacketBatch(total, data []byte, direct bool) ([]byte, error) {
	if data == nil {
		//only flush the buffer
		if direct == true {
			n, err := p.wb.Write(total)
			if err != nil {
				return nil, ErrBadConn
			}
			if n != len(total) {
				return nil, ErrBadConn
			}
		}
		return total, nil
	}

	length := len(data) - 4
	for length >= MaxPayloadLen {

		data[0] = 0xff
		data[1] = 0xff
		data[2] = 0xff

		data[3] = p.Sequence
		total = append(total, data[:4+MaxPayloadLen]...)

		p.Sequence++
		length -= MaxPayloadLen
		data = data[MaxPayloadLen:]
	}

	data[0] = byte(length)
	data[1] = byte(length >> 8)
	data[2] = byte(length >> 16)
	data[3] = p.Sequence

	total = append(total, data...)
	p.Sequence++

	if direct {
		if n, err := p.wb.Write(total); err != nil {
			return nil, ErrBadConn
		} else if n != len(total) {
			return nil, ErrBadConn
		}
	}
	return total, nil
}
