package connfig

import (
	"fmt"
)

/*
系统配置信息
*/
type Config struct {
	Addr string
}

func TellConfig() {
	fmt.Println("this is configure file module!")

}
