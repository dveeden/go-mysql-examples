package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-mysql-org/go-mysql/driver"
)

func main() {
	db, err := sql.Open("mysql", "root@127.0.0.1:3307/test")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	var version string
	db.QueryRow("SELECT VERSION()").Scan(&version)
	fmt.Println(version)
}
