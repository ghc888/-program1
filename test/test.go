package main

import (
	"fmt"
	"net"
	"program1/module"
)

const (
	UP_ACTION             = 1 //上传
	DOWN_ACTION           = 2 //下载
	OTHER_ACTION          = 4 //其他
	defaultReaderSize     = 8 * 1024
	MaxPayloadLen     int = 1<<24 - 1
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("connection server error:", err)
	}
	var ServerVersion string = "1.1.2"
	client := module.NewConn(conn)

	//缓冲区 预留4个字节的包头位置,在WritePacket中进行封装包头
	data := make([]byte, 4, 128)

	data = append(data, 1)
	data = append(data, ServerVersion...)

	//发送
	client.WritePacket(data)

}
