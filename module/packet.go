package module

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

func NewPacket() *Packet {
	p := new(Packet)

	return p
}
