package main

import (
	"encoding/gob"
	"fmt"
	"net"
)

/*
坐标信息
*/
type P struct {
	X, Y float32
}

const (
	UP_ACTION    = 1 //上传
	DOWN_ACTION  = 2 //下载
	OTHER_ACTION = 4 //其他
)

/*
数据包格式
*/
type Packet struct {
	Header      uint8  //包头
	HeaderLengh int32  //包长
	Data        []byte //body
}

func main() {

	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("connection server error:", err)
	}

	//发送请求的message
	encoder := gob.NewEncoder(conn)
	p := &P{111.23, 34.3}
	encoder.Encode(p)
	conn.Close()
	fmt.Println("send message ok!", p)
}
