package cdc_event

import (
	"github.com/go-mysql-org/go-mysql/schema"
)

// taken from https://github.com/go-mysql-org/go-mysql-elasticsearch/blob/master/river/sync.go#L273
func getColValue(col *schema.TableColumn, value interface{}) interface{} {
	switch col.Type {
	case schema.TYPE_STRING:
		switch value := value.(type) {
		case []byte:
			return string(value[:])
		}
	}

	return value
}
