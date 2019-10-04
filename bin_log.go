package mysql_binlog

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/eapache/channels"
	"github.com/siddontang/go-mysql/canal"
)

const (
	DefaultActionChanSize uint16 = 500
)

type BinLogService struct {
	logger         *log.Logger
	conf           *Config
	actionChanSize *uint16
	queueDirection QueueDirection

	runOk            bool
	ringLogQueue     *channels.RingChannel
	overflowingQueue *channels.OverflowingChannel
	gormModels       []interface{}
}

func NewBilLogService(options ...func(*BinLogService) error) (*BinLogService, error) {
	service := &BinLogService{}
	for _, op := range options {
		err := op(service)
		if err != nil {
			return nil, err
		}
	}
	return service, nil
}

func SetLogger(logger *log.Logger) func(*BinLogService) error {
	return func(service *BinLogService) error {
		service.logger = logger
		return nil
	}
}

func SetConfig(configs *Config) func(*BinLogService) error {
	return func(service *BinLogService) error {
		if configs == nil {
			return fmt.Errorf("empty configs value")
		}
		service.conf = configs
		return nil
	}
}

func SetQueueDirection(queueDirection QueueDirection) func(*BinLogService) error {
	return func(service *BinLogService) error {
		service.queueDirection = queueDirection
		return nil
	}
}

func SetQueueSize(size *uint16) func(*BinLogService) error {
	return func(service *BinLogService) error {
		service.actionChanSize = size
		return nil
	}
}

func (b *BinLogService) Listen() {
	c, err := b.createCanal()
	if err == nil {
		coords, err := c.GetMasterPos()
		if err == nil {
			handler := binlogHandler{
				service: b,
			}
			if b.queueDirection == Ring {
				if b.actionChanSize != nil {
					b.ringLogQueue = channels.NewRingChannel(channels.BufferCap(*b.actionChanSize))
				} else {
					b.ringLogQueue = channels.NewRingChannel(channels.BufferCap(DefaultActionChanSize))
				}
			} else {
				if b.actionChanSize != nil {
					b.overflowingQueue = channels.NewOverflowingChannel(channels.BufferCap(*b.actionChanSize))
				} else {
					b.overflowingQueue = channels.NewOverflowingChannel(channels.BufferCap(DefaultActionChanSize))
				}
			}
			b.runOk = true
			c.SetEventHandler(&handler)
			c.RunFrom(coords)
		}
	}
	defer c.Close()
	if b.logger != nil {
		b.logger.Fatalf("mysql_binlog_:_Listen_: error: %+v, stack: %s\n", err, string(debug.Stack()))
	} else {
		log.Fatalf("mysql_binlog_:_Listen_: error: %+v, stack: %s\n", err, string(debug.Stack()))
	}
}

func (b *BinLogService) Close() {
	if b.queueDirection == Ring {
		b.ringLogQueue.Close()
	} else {
		b.overflowingQueue.Close()
	}
}

func (b *BinLogService) Subscriber() <-chan interface{} {
	for {
		if b.ringLogQueue == nil && b.overflowingQueue == nil {
			continue
		}
		if b.ringLogQueue != nil {
			return b.ringLogQueue.Out()
		} else {
			return b.overflowingQueue.Out()
		}
	}
}

func (b *BinLogService) createCanal() (*canal.Canal, error) {
	cfg := canal.NewDefaultConfig()
	cfg.Addr = fmt.Sprintf("%s:%d",
		*b.conf.Host,
		*b.conf.Port,
	)
	cfg.User = *b.conf.User
	cfg.Password = *b.conf.Password
	// We only care table 'trade' in test db
	cfg.Dump.TableDB = *b.conf.Schema
	cfg.Dump.ExecutionPath = ""
	cfg.Flavor = "mysql"
	cfg.Dump.Protocol = "udp"
	cfg.MaxReconnectAttempts = 5
	return canal.NewCanal(cfg)
}
