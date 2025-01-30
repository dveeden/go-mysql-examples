package main

import (
	"fmt"

	"github.com/go-mysql-org/go-mysql/client"
)

func main() {
	conn, err := client.Connect("127.0.0.1:3307", "root", "", "test")
	if err != nil {
		panic(err)
	}
	defer conn.Quit()

	fmt.Println(conn.GetServerVersion())
}
