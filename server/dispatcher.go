package server

import (
	"fmt"
	"program1/mysql"
	"program1/mysql/sqlparser"
	"program1/utils"
	"runtime"
	"strings"
)

type ExecuteDB struct {
	// ExecNode *backend.Node
	IsSlave bool
	sql     string
}

func (c *ClientConn) DispatchMessage(message []byte) error {

	command := message[0]
	data := message[1:]

	switch command {
	case mysql.COM_QUIT:
		fmt.Println("revice com_quit command")
		// c.handleRollback()
		c.Close()
		return nil
	case mysql.COM_QUERY:
		fmt.Println("revice com_query command")
		return c.handleQuery(utils.String(data))
	case mysql.COM_PING:
		return c.writeOK(nil)
	case mysql.COM_INIT_DB:
		fmt.Println("revice com_init_db command")
		// return c.handleUseDB(mysql.String(data))
	case mysql.COM_FIELD_LIST:
		fmt.Println("revice com_field_list command")
		// return c.handleFieldList(data)
	case mysql.COM_STMT_PREPARE:
		fmt.Println("revice com_stmt_prepare command")
		// return c.handleStmtPrepare(mysql.String(data))
	case mysql.COM_STMT_EXECUTE:
		fmt.Println("revice com_stmt_execute command")
		// return c.handleStmtExecute(data)
	case mysql.COM_STMT_CLOSE:
		fmt.Println("revice com_stmt_close command")
		// return c.handleStmtClose(data)
	case mysql.COM_STMT_SEND_LONG_DATA:
		fmt.Println("revice com_stmt_send_long_data command")
		// return c.handleStmtSendLongData(data)
	case mysql.COM_STMT_RESET:
		fmt.Println("revice com_stmt_reset command")
		// return c.handleStmtReset(data)
	case mysql.COM_SET_OPTION:
		fmt.Println("revice com_set_option command")
		return c.writeEOF(0)
	default:
		fmt.Println("other message")
		// msg := fmt.Sprintf("command %d not supported now", cmd)
		// golog.Error("ClientConn", "dispatch", msg, 0)
		// return mysql.NewError(mysql.ER_UNKNOWN_ERROR, msg)
	}

	return nil
}

/*处理query语句*/
func (c *ClientConn) handleQuery(sql string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if err, ok := e.(error); ok {
				const size = 4096
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				fmt.Println(err.Error())
			}
			return
		}
	}()

	sql = strings.TrimRight(sql, ";") //删除sql语句最后的分号
	fmt.Println("sql:", sql)
	hasHandled, err := c.preHandleShard(sql)
	if err != nil {
		// golog.Error("server", "preHandleShard", err.Error(), 0,
		// "sql", sql,
		// "hasHandled", hasHandled,
		// )
		return err
	}
	if hasHandled {
		return nil
	}

	var stmt sqlparser.Statement
	stmt, err = sqlparser.Parse(sql) //解析sql语句,得到的stmt是一个interface
	if err != nil {
		// golog.Error("server", "parse", err.Error(), 0, "hasHandled", hasHandled, "sql", sql)
		return err
	}

	switch v := stmt.(type) {
	case *sqlparser.Select:
		fmt.Println(v)
		//return c.handleSelect(v, nil)
	case *sqlparser.Insert:
		//return c.handleExec(stmt, nil)
	case *sqlparser.Update:
		//return c.handleExec(stmt, nil)
	case *sqlparser.Delete:
		//return c.handleExec(stmt, nil)
	case *sqlparser.Replace:
		//return c.handleExec(stmt, nil)
	case *sqlparser.Set:
		//return c.handleSet(v, sql)
	case *sqlparser.Begin:
		//return c.handleBegin()
	case *sqlparser.Commit:
		//return c.handleCommit()
	case *sqlparser.Rollback:
		//return c.handleRollback()
	case *sqlparser.Admin:
		//return c.handleAdmin(v)
	case *sqlparser.AdminHelp:
		//return c.handleAdminHelp(v)
	case *sqlparser.UseDB:
		fmt.Println(v)
		//return c.handleUseDB(v.DB)
	case *sqlparser.SimpleSelect:
		//return c.handleSimpleSelect(v)
	case *sqlparser.Truncate:
		//return c.handleExec(stmt, nil)
	default:
		return fmt.Errorf("statement %T not support now", stmt)
	}

	return nil
}

//preprocessing sql before parse sql
func (c *ClientConn) preHandleShard(sql string) (bool, error) {
	//var rs []*mysql.Result
	var err error
	var executeDB *ExecuteDB

	if len(sql) == 0 {
		return false, fmt.Errorf("unsport command")
	}
	//TODOsql白名单处理

	tokens := strings.FieldsFunc(sql, utils.IsSqlSep)
	fmt.Println(tokens)
	if len(tokens) == 0 {
		return false, fmt.Errorf("unsport comand")
	}

	// if c.isInTransaction() {
	// 	executeDB, err = c.GetTransExecDB(tokens, sql)
	// } else {
	executeDB, err = c.GetExecDB(tokens, sql)
	// }

	if err != nil {
		//this SQL doesn't need execute in the backend.
		// if err == errors.ErrIgnoreSQL {
		err = c.writeOK(nil)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, err
	// }
	//need shard sql
	if executeDB == nil {
		return false, nil
	}
	//get connection in DB
	//conn, err := c.getBackendConn(executeDB.ExecNode, executeDB.IsSlave)
	//defer c.closeConn(conn, false)
	//if err != nil {
	//	return false, err
	//}
	// //execute.sql may be rewritten in getShowExecDB
	// rs, err = c.executeInNode(conn, executeDB.sql, nil)
	// if err != nil {
	// 	return false, err
	// }

	// if len(rs) == 0 {
	// 	msg := fmt.Sprintf("result is empty")
	// 	golog.Error("ClientConn", "handleUnsupport", msg, 0, "sql", sql)
	// 	return false, mysql.NewError(mysql.ER_UNKNOWN_ERROR, msg)
	// }

	// c.lastInsertId = int64(rs[0].InsertId)
	// c.affectedRows = int64(rs[0].AffectedRows)

	// if rs[0].Resultset != nil {
	// 	err = c.writeResultset(c.status, rs[0].Resultset)
	// } else {
	// 	err = c.writeOK(rs[0])
	// }

	// if err != nil {
	// 	return false, err
	// }

	return true, nil
}

//if sql need shard return nil, else return the unshard db
func (c *ClientConn) GetExecDB(tokens []string, sql string) (*ExecuteDB, error) {
	tokensLen := len(tokens)
	if 0 < tokensLen {
		tokenId, ok := mysql.PARSE_TOKEN_MAP[strings.ToLower(tokens[0])]
		if ok == true {
			switch tokenId {
			case mysql.TK_ID_SELECT:
				fmt.Println("select command ")
				//return c.getSelectExecDB(sql, tokens, tokensLen)
			case mysql.TK_ID_DELETE:
				fmt.Println("delete command ")
				//return c.getDeleteExecDB(sql, tokens, tokensLen)
			case mysql.TK_ID_INSERT, mysql.TK_ID_REPLACE:
				fmt.Println("insert repliace command")
				//return c.getInsertOrReplaceExecDB(sql, tokens, tokensLen)
			case mysql.TK_ID_UPDATE:
				fmt.Println("update command")
				//return c.getUpdateExecDB(sql, tokens, tokensLen)
			case mysql.TK_ID_SET:
				fmt.Println("set commmand")
				//return c.getSetExecDB(sql, tokens, tokensLen)
			case mysql.TK_ID_SHOW:
				fmt.Println("show command")
				//return c.getShowExecDB(sql, tokens, tokensLen)
			case mysql.TK_ID_TRUNCATE:
				fmt.Println("truncate command ")
				//return c.getTruncateExecDB(sql, tokens, tokensLen)
			default:
				return nil, nil
			}
		}
	}
	executeDB := new(ExecuteDB)
	executeDB.sql = sql
	//err := c.setExecuteNode(tokens, tokensLen, executeDB)
	// if err != nil {
	// 	return nil, err
	// }
	return executeDB, nil
}
