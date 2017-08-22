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
