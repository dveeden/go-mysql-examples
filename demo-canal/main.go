package main

import (
	"fmt"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
)

type MyEventHandler struct {
	canal.DummyEventHandler
}

func (h *MyEventHandler) OnRow(e *canal.RowsEvent) error {
	for _, r := range e.Rows {
		fmt.Printf("action=%s first_col=%#v\n", e.Action, r[0])
	}
	return nil
}

func main() {
	cfg := canal.NewDefaultConfig()
	cfg.Addr = "127.0.0.1:3307"
	cfg.User = "root"
	cfg.Dump.TableDB = "test"
	cfg.Dump.Tables = []string{"t"}

	c, _ := canal.NewCanal(cfg)
	c.SetEventHandler(&MyEventHandler{})
	gtid, _ := mysql.ParseGTIDSet(mysql.MySQLFlavor,
		"94e1b69b-d8df-11ef-85d2-c6bc6abebafa:1")
	c.StartFromGTID(gtid)
}
