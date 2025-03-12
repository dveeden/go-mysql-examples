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
	cfg.Addr = "127.0.0.1:3306"
	cfg.User = "root"
	cfg.Dump.ExecutionPath = ""
	c, _ := canal.NewCanal(cfg)
	c.SetEventHandler(&MyEventHandler{})
	gtid, _ := mysql.ParseGTIDSet(mysql.MySQLFlavor,
		"896e7882-18fe-11ef-ab88-22222d34d411:1")
	c.StartFromGTID(gtid)
}
