package main

import (
	"context"
	"fmt"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
)

func main() {
	cfg := replication.BinlogSyncerConfig{
		ServerID: 123,
		Flavor:   "mysql",
		Host:     "127.0.0.1",
		Port:     3306,
		User:     "root",
		Password: "",
	}
	syncer := replication.NewBinlogSyncer(cfg)
	streamer, _ := syncer.StartSync(mysql.Position{"binlog.000003", 219532})
	for {
		ev, _ := streamer.GetEvent(context.Background())

		if e, ok := ev.Event.(*replication.RowsEvent); ok {
			for _, r := range e.Rows {
				fmt.Printf("value of first column: %d\n", r[0])
			}
		}
	}
}
