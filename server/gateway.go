package server

import (
	"fmt"
	"program1/config"
)

/*
入口server
*/
type GateServer struct {
}

func GRun() {
	fmt.Println("hello world!")
	connfig.TellConfig()
}
