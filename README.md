# mysql_binlog

## Installation

Run command on you [$GOPATH/src] path:

```bash
go get -u github.com/alex60217101990/mysql_binlog
```

## Usage

[...]

Create model:

```go
type GameLogo struct {
	ID          int64  `json:"-" gorm:"primary_key;column:id;" sql:"type:int unsigned auto_increment"`
	GameID      int64  `json:"game_id" gorm:"column:game_id;type:bigint;"`
	GameLogoURL string `json:"game_logo" gorm:"type:varchar(500);column:game_logo;"`
}

// must be implement method: [TableName() string]
func (GameLogo) TableName() string {
	return "games_logos"
}

// must be implement method: [FromBinLog(fields ...interface{}) interface{}]
func (GameLogo) FromBinLog(fields ...interface{}) interface{} {
	g := GameLogo{}
	g.ID = int64(fields[0].(int32))
	g.GameID = fields[1].(int64)
	g.GameLogoURL = fields[2].(string)
	return g
}
```
then:

```go
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
			GormModels: []DBModel{&GameLogo{}}, // or other you gorm models
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
```
[...]

'BinLogEvent' has signature:

[...]
```go
type BinLogEvent struct {
	Type     ActionType
	Value    interface{}
	OldValue interface{}
}
```
[...]

## License
[MIT](https://choosealicense.com/licenses/mit/)