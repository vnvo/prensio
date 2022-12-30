package mysql_source

import (
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	"github.com/siddontang/go-log/log"
	cdc "github.com/vnvo/prensio/pipeline/cdc_event"
)

type eventHandler struct {
	source   *MySQLBinlogSource
	lastGtid *mysql.GTIDSet
}

func (h *eventHandler) OnRotate(e *replication.RotateEvent) error {
	pos := mysql.Position{
		Name: string(e.NextLogName),
		Pos:  uint32(e.Position),
	}

	log.Debugf("OnRotate. Name: %s, Position: %d", pos.Name, pos.Pos)
	return nil
}

func (h *eventHandler) OnTableChanged(schema, table string) error {
	log.Debugf("OnTableChanged. table: %s, schema: %s", schema, table)
	return nil
}

func (h *eventHandler) OnDDL(nextPos mysql.Position, _ *replication.QueryEvent) error {
	log.Debugf("OnDLL. nextPos: %s", nextPos)
	return nil
}

func (h *eventHandler) OnXID(nextPos mysql.Position) error {
	log.Debugf("OnXID. nextPosition.Name: %s, nextPosition.Pos: %d", nextPos.Name, nextPos.Pos)
	return nil
}

func (h *eventHandler) OnRow(e *canal.RowsEvent) error {
	h.source.eventCh <- cdc.NewCDCEvent(e, h.lastGtid)
	return nil
}

func (h *eventHandler) OnGTID(gtid mysql.GTIDSet) error {
	log.Debugf("OnGTID. %s", gtid.String())
	h.lastGtid = &gtid
	return nil
}

func (h *eventHandler) String() string {
	return "DefaultMySQLEventHandler"
}

func (h *eventHandler) OnPosSynced(pos mysql.Position, set mysql.GTIDSet, force bool) error {
	log.Debugf("OnPosSynced. pos=%v, set=%v, force=%v", pos, set, force)
	return nil
}
