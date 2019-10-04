package mysql_binlog

import (
	"log"
	"runtime/debug"

	"github.com/siddontang/go-mysql/canal"
)

type binlogHandler struct {
	service                 *BinLogService
	canal.DummyEventHandler // Standard stock from the library
	BinlogParser            // Our custom helper for processing
}

func (h *binlogHandler) OnRow(e *canal.RowsEvent) error {
	defer func() {
		if r := recover(); r != nil {
			if h.service.logger != nil {
				h.service.logger.Printf("mysql_binlog_:_OnRow_: fatal error: %+v, stack: %s\n", r.(error), string(debug.Stack()))
			} else {
				log.Printf("mysql_binlog_:_OnRow_: fatal error: %+v, stack: %s", r.(error), string(debug.Stack()))
			}
		}
	}()
	// base value for canal.DeleteAction or canal.InsertAction
	var n = 0
	var k = 1
	if e.Action == canal.UpdateAction {
		n = 1
		k = 2
	}

	for i := n; i < len(e.Rows); i += k {
	Loop:
		for _, model := range h.service.conf.GormModels {
			key := e.Table.Schema + "." + e.Table.Name
			tableName, err := parseMethodCall(model, "TableName")
			if err != nil {
				if h.service.logger != nil {
					h.service.logger.Printf("mysql_binlog_:_OnRow_: error: %+v, stack: %s\n", err, string(debug.Stack()))
				} else {
					log.Printf("mysql_binlog_:_OnRow_: error: %+v, stack: %s", err, string(debug.Stack()))
				}
				break Loop
			}
			switch key {
			case *h.service.conf.Schema + "." + tableName.String():
				newModel, err := parseMethodCall(model, "FromBinLog", e.Rows[i]...)
				if err != nil {
					if h.service.logger != nil {
						h.service.logger.Printf("mysql_binlog_:_OnRow_: error: %+v, stack: %s\n", err, string(debug.Stack()))
					} else {
						log.Printf("mysql_binlog_:_OnRow_: error: %+v, stack: %s", err, string(debug.Stack()))
					}
					break Loop
				}
				event := BinLogEvent{
					Value: newModel.Interface(),
				}
				switch e.Action {
				case canal.UpdateAction:
					oldModel, err := parseMethodCall(model, "FromBinLog", e.Rows[i-1]...)
					if err != nil {
						if h.service.logger != nil {
							h.service.logger.Printf("mysql_binlog_:_OnRow_: error: %+v, stack: %s\n", err, string(debug.Stack()))
						} else {
							log.Printf("mysql_binlog_:_OnRow_: error: %+v, stack: %s", err, string(debug.Stack()))
						}
						break Loop
					}
					event.Type = Update
					event.OldValue = oldModel.Interface()
				case canal.InsertAction:
					event.Type = Insert
				case canal.DeleteAction:
					event.Type = Delete
				default:
					if h.service.logger != nil {
						h.service.logger.Printf("mysql_binlog_:_OnRow_: error: unknown action, stack: %s\n", string(debug.Stack()))
					} else {
						log.Printf("mysql_binlog_:_OnRow_: error: unknown action, stack: %s\n", string(debug.Stack()))
					}
					break Loop
				}
				h.emmitBinLogEvent(event)
			}
		}
	}
	return nil
}

func (h *binlogHandler) String() string { return "binlogHandler" }

func (h *binlogHandler) emmitBinLogEvent(event interface{}) {
	if h.service.ringLogQueue != nil {
		// custom_logger.CmdServer.Println(event.(BinLogEvent))
		h.service.ringLogQueue.In() <- event
	} else {
		h.service.overflowingQueue.In() <- event
	}
}
