package mysql

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

/*
数据对象连接信息
*/

const (
	Up = iota
	Down
	ManualDown
	Unknown

	InitConnCount           = 16
	DefaultMaxConnNum       = 1024
	PingPeroid        int64 = 4
)

type DB struct {
	sync.RWMutex

	addr     string
	user     string
	password string
	db       string
	state    int32

	maxConnNum  int
	InitConnNum int
	idleConns   chan *DBConn
	cacheConns  chan *DBConn
	checkConn   *DBConn
	lastPing    int64
	Is_alive    bool
}

func Open(addr string, user string, password string, dbName string, maxConnNum int) (*DB, error) {
	var err error
	db := new(DB)
	db.addr = addr
	db.user = user
	db.password = password
	db.db = dbName

	if 0 < maxConnNum {
		db.maxConnNum = maxConnNum
		if db.maxConnNum < 16 {
			db.InitConnNum = db.maxConnNum
		} else {
			db.InitConnNum = db.maxConnNum / 4
		}
	} else {
		db.maxConnNum = DefaultMaxConnNum
		db.InitConnNum = InitConnCount
	}
	//check connection
	db.checkConn, err = db.newConn()
	if err != nil {
		db.Close()
		return nil, err
	}

	db.idleConns = make(chan *DBConn, db.maxConnNum)
	db.cacheConns = make(chan *DBConn, db.maxConnNum)
	atomic.StoreInt32(&(db.state), Unknown)

	for i := 0; i < db.maxConnNum; i++ {
		if i < db.InitConnNum {
			conn, err := db.newConn()
			if err != nil {
				db.Close()
				return nil, err
			}
			conn.pushTimestamp = time.Now().Unix()
			db.cacheConns <- conn
		} else {
			conn := new(DBConn)
			db.idleConns <- conn
		}
	}
	db.SetLastPing()

	return db, nil
}

func (db *DB) Addr() string {
	return db.addr
}

func (db *DB) State() string {
	var state string
	switch db.state {
	case Up:
		state = "up"
	case Down, ManualDown:
		state = "down"
	case Unknown:
		state = "unknow"
	}
	return state
}

func (db *DB) IdleConnCount() int {
	db.RLock()
	defer db.RUnlock()
	return len(db.cacheConns)
}

func (db *DB) Close() error {
	db.Lock()
	idleChannel := db.idleConns
	cacheChannel := db.cacheConns
	db.cacheConns = nil
	db.idleConns = nil
	db.Unlock()
	if cacheChannel == nil || idleChannel == nil {
		return nil
	}

	close(cacheChannel)
	for conn := range cacheChannel {
		db.closeConn(conn)
	}
	close(idleChannel)

	return nil
}

func (db *DB) getConns() (chan *DBConn, chan *DBConn) {
	db.RLock()
	cacheConns := db.cacheConns
	idleConns := db.idleConns
	db.RUnlock()
	return cacheConns, idleConns
}

func (db *DB) getCacheConns() chan *DBConn {
	db.RLock()
	conns := db.cacheConns
	db.RUnlock()
	return conns
}

func (db *DB) getIdleConns() chan *DBConn {
	db.RLock()
	conns := db.idleConns
	db.RUnlock()
	return conns
}

func (db *DB) Ping() error {
	var err error
	if db.checkConn == nil {
		db.checkConn, err = db.newConn()
		if err != nil {
			db.Is_alive = false
			db.closeConn(db.checkConn)
			db.checkConn = nil
			return err
		}
	}
	err = db.checkConn.Ping()
	if err != nil {
		db.Is_alive = false
		db.closeConn(db.checkConn)
		db.checkConn = nil
		return err
	}
	db.Is_alive = true
	return nil
}

func (db *DB) newConn() (*DBConn, error) {
	co := new(DBConn)
	if err := co.Connect(db.addr, db.user, db.password, db.db); err != nil {
		db.Is_alive = false
		return nil, err
	}
	return co, nil
}

func (db *DB) closeConn(co *DBConn) error {
	if co != nil {
		co.Close()
		conns := db.getIdleConns()
		if conns != nil {
			select {
			case conns <- co:
				return nil
			default:
				return nil
			}
		}
	}
	return nil
}

func (db *DB) tryReuse(co *DBConn) error {
	var err error
	//reuse Connection
	if co.IsInTransaction() {
		//we can not reuse a connection in transaction status
		err = co.Rollback()
		if err != nil {
			return err
		}
	}

	if !co.IsAutoCommit() {
		//we can not  reuse a connection not in autocomit
		_, err = co.exec("set autocommit = 1")
		if err != nil {
			return err
		}
	}

	//connection may be set names early
	//we must use default utf8
	if co.GetCharset() != DEFAULT_CHARSET {
		err = co.SetCharset(DEFAULT_CHARSET, DEFAULT_COLLATION_ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) PopConn() (*DBConn, error) {
	var co *DBConn
	var err error

	cacheConns, idleConns := db.getConns()
	if cacheConns == nil || idleConns == nil {
		return nil, fmt.Errorf("database close")
	}
	co = db.GetConnFromCache(cacheConns)
	if co == nil {
		co, err = db.GetConnFromIdle(cacheConns, idleConns)
		if err != nil {
			return nil, err
		}
	}

	err = db.tryReuse(co)
	if err != nil {
		db.closeConn(co)
		return nil, err
	}

	return co, nil
}

func (db *DB) GetConnFromCache(cacheConns chan *DBConn) *DBConn {
	var co *DBConn
	var err error
	for 0 < len(cacheConns) {
		co = <-cacheConns
		if co != nil && PingPeroid < time.Now().Unix()-co.pushTimestamp {
			err = co.Ping()
			if err != nil {
				db.closeConn(co)
				co = nil
			}
		}
		if co != nil {
			break
		}
	}
	return co
}

func (db *DB) GetConnFromIdle(cacheConns, idleConns chan *DBConn) (*DBConn, error) {
	var co *DBConn
	var err error
	select {
	case co = <-idleConns:
		err = co.Connect(db.addr, db.user, db.password, db.db)
		if err != nil {
			db.closeConn(co)
			return nil, err
		}
		return co, nil
	case co = <-cacheConns:
		if co == nil {
			return nil, fmt.Errorf("conn is nil")
		}
		if co != nil && PingPeroid < time.Now().Unix()-co.pushTimestamp {
			err = co.Ping()
			if err != nil {
				db.closeConn(co)
				return nil, fmt.Errorf("bad conn")
			}
		}
	}
	return co, nil
}

func (db *DB) PushConn(co *DBConn, err error) {
	if co == nil {
		return
	}
	conns := db.getCacheConns()
	if conns == nil {
		co.Close()
		return
	}
	if err != nil {
		db.closeConn(co)
		return
	}
	co.pushTimestamp = time.Now().Unix()
	select {
	case conns <- co:
		return
	default:
		db.closeConn(co)
		return
	}
}

type BackendConn struct {
	*DBConn
	db *DB
}

func (p *BackendConn) Close() {
	if p != nil && p.DBConn != nil {
		if p.DBConn.pkgErr != nil {
			p.db.closeConn(p.DBConn)
		} else {
			p.db.PushConn(p.DBConn, nil)
		}
		p.DBConn = nil
	}
}

func (db *DB) GetConn() (*BackendConn, error) {
	c, err := db.PopConn()
	if err != nil {
		return nil, err
	}
	return &BackendConn{c, db}, nil
}

func (db *DB) SetLastPing() {
	db.lastPing = time.Now().Unix()
}

func (db *DB) GetLastPing() int64 {
	return db.lastPing
}
