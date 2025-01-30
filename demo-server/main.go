package main

import (
	"net"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/server"
)

type DemoHandler struct {
	server.EmptyHandler
}

func (h DemoHandler) HandleQuery(query string) (*mysql.Result, error) {
	if query == `SELECT 2+2` {
		r, _ := mysql.BuildSimpleResultset(
			[]string{"result"},
			[][]interface{}{{"5"}}, false)
		return mysql.NewResult(r), nil
	}
	return nil, nil
}

func main() {
	l, _ := net.Listen("tcp", "127.0.0.1:4000")
	c, _ := l.Accept()
	conn, _ := server.NewConn(c, "root", "", DemoHandler{})
	for {
		if err := conn.HandleCommand(); err != nil {
			panic(err)
		}
	}
}
