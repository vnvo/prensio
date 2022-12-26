package cdc_event

import (
	"encoding/json"
	"time"

	"github.com/go-mysql-org/go-mysql/canal"
)

type CDCEvent struct {
	raw    *canal.RowsEvent
	Schema string                   `json:"schema"`
	Table  string                   `json:"table"`
	Action string                   `json:"action"`
	Before []map[string]interface{} `json:"before"`
	After  []map[string]interface{} `json:"after"`
	Meta   CDCEventMeta             `json:"meta"`
}

type CDCEventMeta struct {
	Timestamp int64  `json:"timestamp"`
	Pipeline  string `json:"pipeline"`
}

func NewCDCEvent(rawEvent *canal.RowsEvent) CDCEvent {
	cdcEvent := CDCEvent{
		rawEvent,
		rawEvent.Table.Schema,
		rawEvent.Table.Name,
		rawEvent.Action,
		nil,
		nil,
		CDCEventMeta{time.Now().UnixMicro(), "test-pipeline"},
	}

	switch cdcEvent.Action {
	case "insert":
		cdcEvent.handleInsert()
	case "update":
		cdcEvent.handleUpdate()
	case "delete":
		cdcEvent.handleDelete()
	}

	return cdcEvent
}

func (e *CDCEvent) handleInsert() {

	e.After = make([]map[string]interface{}, len(e.raw.Rows))

	for i, row := range e.raw.Rows {
		e.After[i] = make(map[string]interface{})

		for col_idx := range e.raw.Table.Columns {
			col := e.raw.Table.Columns[col_idx]
			e.After[i][col.Name] = getColValue(&col, row[col_idx])
		}

	}

}

func (e *CDCEvent) handleDelete() {

	e.Before = make([]map[string]interface{}, len(e.raw.Rows))

	for i, row := range e.raw.Rows {
		e.Before[i] = make(map[string]interface{})

		for col_idx := range e.raw.Table.Columns {
			col := e.raw.Table.Columns[col_idx]
			e.Before[i][col.Name] = getColValue(&col, row[col_idx])
		}

	}

}

func (e *CDCEvent) handleUpdate() {

	rows := e.raw.Rows

	e.Before = make([]map[string]interface{}, 0)
	e.After = make([]map[string]interface{}, 0)

	for i := 0; i < len(rows); i += 2 {
		before := make(map[string]interface{})
		after := make(map[string]interface{})

		for col_idx := range e.raw.Table.Columns {
			col := e.raw.Table.Columns[col_idx]

			before[col.Name] = getColValue(&col, rows[i][col_idx])
			after[col.Name] = getColValue(&col, rows[i+1][col_idx])
		}

		e.Before = append(e.Before, before)
		e.After = append(e.After, after)

	}
}

func (e *CDCEvent) ToJson() (string, error) {
	ej, err := json.Marshal(e)

	if err != nil {
		return "", err
	}

	return string(ej), nil
}
