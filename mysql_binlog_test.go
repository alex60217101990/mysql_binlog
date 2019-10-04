package mysql_binlog

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestBinLog(t *testing.T) {
	host := "localhost"
	port := 3307
	schema := "trade"
	user := "root"
	password := "password"
	size := uint16(200)

	binLog, err := NewBilLogService(
		SetQueueDirection(Ring),
		SetQueueSize(&size),
		SetConfig(&Config{
			Host:       &host,
			Port:       &port,
			Schema:     &schema,
			User:       &user,
			Password:   &password,
			GormModels: []DBModel{&GameLogo{}},
		}),
	)
	if err != nil {
		log.Println(err)
	}

	go binLog.Listen()

	defer binLog.Close()

	go func() {
		for {
			select {
			case event, ok := <-binLog.Subscriber():
				if ok {
					if logEvent, ok := event.(BinLogEvent); ok {
						if logo, ok := logEvent.Value.(GameLogo); ok {
							fmt.Println(logo)
						}
						if logEvent.OldValue != nil {
							if logo, ok := logEvent.OldValue.(GameLogo); ok {
								fmt.Println(logo)
							}
						}
					}
				}
			}
		}
	}()

	time.Sleep(30 * time.Second)
}
