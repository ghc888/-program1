package server

import (
	"fmt"
	"program1/mysql"
)

func (c *ClientConn) DispatchMessage(message []byte) error {

	command := message[0]
	data := message[1:]
	fmt.Println(string(data))
	switch command {
	case mysql.COM_QUIT:
		fmt.Println("revice com_quit command")
		// c.handleRollback()
		// c.Close()
		// return nil
	case mysql.COM_QUERY:
		fmt.Println("revice com_query command")
		// return c.handleQuery(mysql.String(data))
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
