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

	res, _ := conn.Execute("SELECT VERSION() AS ver")
	defer res.Close()

	version, _ := res.GetStringByName(0, "ver")
	fmt.Println(version)
}
