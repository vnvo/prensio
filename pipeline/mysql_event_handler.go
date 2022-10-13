package pipeline

import (
	"fmt"

	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
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
	fmt.Println("OnRow. event:", e)
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
	return nil
}

func (mys *MySQLBinlogSource) handlerLoop() {
}
