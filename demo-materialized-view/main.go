package main

import (
	"bytes"
	"context"
	"flag"
	"log/slog"
	"net"
	"strconv"

	"github.com/go-mysql-org/go-mysql/client"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
)

type MaterializedView interface {
	HandleInsert(rows [][]interface{}) error
	HandleUpdate(rows [][]interface{}) error
	HandleDelete(rows [][]interface{}) error
	HandleXid() error
	Table() ([]byte, []byte)
}

type MvDemo struct {
	db       string
	table    string
	colnr    uint
	count    uint64
	total    int64
	conn     *client.Conn
	modified bool
}

func NewMvDemo(db, table string, colnr uint) *MvDemo {
	mv := new(MvDemo)
	mv.db = db
	mv.table = table
	mv.colnr = colnr
	return mv
}

func (m *MvDemo) Table() ([]byte, []byte) {
	return []byte(m.db), []byte(m.table)
}

func (m *MvDemo) HandleInsert(rows [][]interface{}) error {
	for _, row := range rows {
		m.count++
		if v, ok := row[m.colnr].(int32); ok {
			m.total += int64(v)
		}
	}
	m.modified = true
	return nil
}
func (m *MvDemo) HandleUpdate(rows [][]interface{}) error {
	for i, row := range rows {
		if v, ok := row[m.colnr].(int32); ok {
			if i%2 == 0 {
				m.total -= int64(v)
			} else {
				m.total += int64(v)
			}
		}
	}
	m.modified = true
	return nil
}
func (m *MvDemo) HandleDelete(ai [][]interface{}) error {
	for _, row := range ai {
		m.count--
		if v, ok := row[m.colnr].(int32); ok {
			m.total -= int64(v)
		}
	}
	return nil
}
func (m *MvDemo) HandleXid() error {
	if !m.modified {
		return nil
	}
	avg := float64(m.total) / float64(m.count)
	slog.Info("updating mv",
		"total", m.total,
		"count", m.count,
		"avg", avg)

	_, err := m.conn.Execute(`REPLACE INTO mv(id, val_avg) VALUES(1, ?)`, avg)
	if err != nil {
		return err
	}
	m.modified = false

	return nil
}

func (m *MvDemo) Initialize(host, user, password string, port int) (mysql.GTIDSet, error) {
	var err error
	m.conn, err = client.Connect(net.JoinHostPort(host, strconv.Itoa(port)), user, password, m.db)
	if err != nil {
		return nil, err
	}

	err = m.conn.Begin()
	if err != nil {
		return nil, err
	}

	avg, err := m.conn.Execute(`SELECT COUNT(val) AS count, SUM(val) AS total FROM t`)
	if err != nil {
		return nil, err
	}

	count, err := avg.GetUint(0, 0)
	if err != nil {
		return nil, err
	}
	total, err := avg.GetInt(0, 1)
	if err != nil {
		return nil, err
	}
	avg.Close()
	m.count = uint64(count)
	m.total = int64(total)

	pos, err := m.conn.Execute(`SELECT @@global.gtid_executed`)
	if err != nil {
		return nil, err
	}
	gtidStr, err := pos.GetString(0, 0)
	if err != nil {
		return nil, err
	}

	gtid, err := mysql.ParseGTIDSet(mysql.MySQLFlavor, gtidStr)
	if err != nil {
		return nil, err
	}

	err = m.conn.Rollback()
	if err != nil {
		return nil, err
	}

	return gtid, nil
}

func handleEvent(ev *replication.BinlogEvent, mv MaterializedView) error {
	var err error
	if e, ok := ev.Event.(*replication.RowsEvent); ok {
		if e.Table != nil {
			s, t := mv.Table()
			if !bytes.EqualFold(e.Table.Schema, s) || !bytes.EqualFold(e.Table.Table, t) {
				return nil
			}

			switch ev.Header.EventType {
			case replication.WRITE_ROWS_EVENTv0,
				replication.WRITE_ROWS_EVENTv1,
				replication.WRITE_ROWS_EVENTv2:
				err = mv.HandleInsert(e.Rows)
			case replication.UPDATE_ROWS_EVENTv0,
				replication.UPDATE_ROWS_EVENTv1,
				replication.UPDATE_ROWS_EVENTv2:
				err = mv.HandleUpdate(e.Rows)
			case replication.DELETE_ROWS_EVENTv0,
				replication.DELETE_ROWS_EVENTv1,
				replication.DELETE_ROWS_EVENTv2:
				err = mv.HandleDelete(e.Rows)
			}
			if err != nil {
				return err
			}
		}
	}
	if _, ok := ev.Event.(*replication.XIDEvent); ok {
		err = mv.HandleXid()
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	var (
		host     = flag.String("h", "127.0.0.1", "hostname")
		port     = flag.Int("P", 3306, "port")
		user     = flag.String("u", "root", "username")
		password = flag.String("p", "", "password")
		mv       MaterializedView
		gtid     mysql.GTIDSet
		err      error
	)
	flag.Parse()
	mv = NewMvDemo("test", "t", 1)
	if mvd, ok := mv.(*MvDemo); ok {
		gtid, err = mvd.Initialize(*host, *user, *password, *port)
		if err != nil {
			panic(err)
		}
		slog.Info("initialized", "gtid", gtid)
	}
	cfg := replication.BinlogSyncerConfig{
		ServerID: 123,
		Flavor:   mysql.MySQLFlavor,
		Host:     *host,
		Port:     uint16(*port),
		User:     *user,
		Password: *password,
	}
	syncer := replication.NewBinlogSyncer(cfg)
	streamer, err := syncer.StartSyncGTID(gtid)
	if err != nil {
		panic(err)
	}
	for {
		ev, err := streamer.GetEvent(context.Background())
		if err != nil {
			panic(err)
		}
		err = handleEvent(ev, mv)
		if err != nil {
			panic(err)
		}
	}
}
