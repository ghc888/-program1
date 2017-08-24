package mysql

import (
	"fmt"
	"sync"
	"time"
)

/*
MySQL 数据节点信息模块
*/

type Node struct {
	Master *DB
	Slave  *DB
	sync.RWMutex
	DownAfterNoAlive time.Duration
}

/*
从Node字节中获取一个数据对象连接 默认
@parm is_slave : 是否从 只读库获取连接
*/
func (n *Node) GetNodeConn(is_slave bool) (*BackendConn, error) {

	if is_slave {
		//读取从库,
		n.Slave.checkConn()

	}
	db := n.Master

	if db == nil {
		return nil, fmt.Errorf("No connection find!")
	}
	// if atomic.LoadInt32(&(db.state)) == Down {
	// 	return nil, errors.ErrMasterDown
	// }

	return db.GetConn()
}