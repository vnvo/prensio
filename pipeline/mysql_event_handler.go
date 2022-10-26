package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	cdc "github.com/vnvo/go-mysql-kafka/cdc_event"
)

type eventHandler struct {
	source *MySQLBinlogSource
}

func (h *eventHandler) OnRotate(e *replication.RotateEvent) error {
	pos := mysql.Position{
		Name: string(e.NextLogName),
		Pos:  uint32(e.Position),
	}

	fmt.Println("OnRotate. Pos:", pos)
	return nil
}

func (h *eventHandler) OnTableChanged(schema, table string) error {
	fmt.Println("OnTableChanged. table, schema:", table, schema)
	return nil
}

func (h *eventHandler) OnDDL(nextPos mysql.Position, _ *replication.QueryEvent) error {
	fmt.Println("OnDLL. nextPos:", nextPos)
	return nil
}

func (h *eventHandler) OnXID(nextPos mysql.Position) error {
	fmt.Println("OnXID. nextPos:", nextPos)
	return nil
}

func (h *eventHandler) OnRow(e *canal.RowsEvent) error {
	h.source.eventCh <- cdc.NewCDCEvent(e)

	return nil
}

func (h *eventHandler) OnGTID(gtid mysql.GTIDSet) error {
	fmt.Println("OnGITD. gtid:", gtid)
	return nil
}

func (h *eventHandler) String() string {
	return "DefaultMySQLEventHandler"
}

func (h *eventHandler) OnPosSynced(pos mysql.Position, set mysql.GTIDSet, force bool) error {
	fmt.Printf("OnPosSynced. pos=%v, set=%v, force=%v\n", pos, set, force)
	return nil
}

func (mys *MySQLBinlogSource) readFromHandler(ctx context.Context, transform string) {
	for {
		select {
		case e := <-mys.eventCh:
			fmt.Println(e.ToJson())
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 1):
			//fmt.Println("just waiting for other channels ...")
		}
	}
}
