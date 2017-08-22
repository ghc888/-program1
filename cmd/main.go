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

func handleConnection(conn net.Conn) {
	dec := gob.NewDecoder(conn)
	p := &P{}
	dec.Decode(p)
	//fmt.Printf("Received : %+v\n", p)
	fmt.Printf("Recive X value is:%d  Y value is:%d", p.X, p.Y)

}

func main() {

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("server listen error:", err)
	}

	for {
		con, err := listener.Accept()
		if err != nil {
			fmt.Println("server accept error:", err)
			continue
		}
		go handleConnection(con)
	}

}
