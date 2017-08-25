package mysql

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

/*
MySQL 数据节点信息模块
*/

type NodeCluster struct {
	Master *DB
	Slave  *DB
	sync.RWMutex
	DownAfterNoAlive time.Duration
}

func test() {
	fmt.Println("hello world!")
}

/*
从Node字节中获取一个数据对象连接 默认
@parm is_slave : 是否从 只读库获取连接
*/
func (n *NodeCluster) GetNodeConn(is_slave bool) (*BackendConn, error) {

	if is_slave {
		//读取从库,
		if n.Slave.Is_alive {
			return n.Slave.GetConn()
		} else {
			return n.Master.GetConn()
		}

	}
	db := n.Master

	if db == nil {
		return nil, fmt.Errorf("No connection find!")
	}
	if atomic.LoadInt32(&(db.state)) == Down {
		return nil, fmt.Errorf("master down")
	}

	return db.GetConn()
}
