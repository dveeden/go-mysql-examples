package main

import (
	"database/sql"
	"log/slog"
	"net"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/server"

	_ "github.com/lib/pq"
)

type SQLHandler struct {
	server.EmptyHandler
	db *sql.DB
}

func (h SQLHandler) HandleQuery(query string) (*mysql.Result, error) {
	slog.Info("HandleQuery", "query", query)

	// These two queries are implemented for minimal support for MySQL Shell
	if query == `SET NAMES 'utf8mb4';` {
		return nil, nil
	}
	if query == `select concat(@@version, ' ', @@version_comment)` {
		r, err := mysql.BuildSimpleResultset([]string{"concat(@@version, ' ', @@version_comment)"}, [][]interface{}{
			{"8.0.11"},
		}, false)
		if err != nil {
			return nil, err
		}
		return mysql.NewResult(r), nil
	}

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if len(columns) == 0 {
		return nil, nil
	}

	data := make([]interface{}, len(columns))
	for i := 0; i < len(columns); i++ {
		var dataValue string
		data[i] = &dataValue
	}
	var res [][]interface{}
	for rows.Next() {
		r := []interface{}{}
		rows.Scan(data...)
		for _, d := range data {
			var colStr string
			if colStrPtr := d.(*string); colStrPtr != nil {
				colStr = *colStrPtr
			}
			r = append(r, colStr)
		}
		res = append(res, r)
	}

	r, _ := mysql.BuildSimpleResultset(
		columns, res, false)
	return mysql.NewResult(r), nil
	return nil, nil
}

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:4000")
	if err != nil {
		panic(err)
	}
	c, err := l.Accept()
	if err != nil {
		panic(err)
	}
	h := SQLHandler{}
	h.db, err = sql.Open("postgres", `host=127.0.0.1 port=5432 user=postgres sslmode=disable`)
	if err != nil {
		panic(err)
	}
	conn, err := server.NewConn(c, "root", "", h)
	if err != nil {
		panic(err)
	}
	for {
		if err := conn.HandleCommand(); err != nil {
			panic(err)
		}
	}
}
