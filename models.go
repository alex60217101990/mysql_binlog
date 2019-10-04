package mysql_binlog

type DBModel interface {
	TableName() string
	FromBinLog(fields ...interface{}) interface{}
}

type Config struct {
	Host       *string
	Port       *int
	Schema     *string
	User       *string
	Password   *string
	GormModels []DBModel
}

type BinLogEvent struct {
	Type     ActionType
	Value    interface{}
	OldValue interface{}
}

type QueueDirection int8

const (
	Ring QueueDirection = (1 + iota)
	Overflowing
)

func (t QueueDirection) String() string {
	names := [...]string{
		"ring queue direction",
		"overflowing queue direction",
	}
	if t < Ring || t > Overflowing {
		return "unknown queue direction"
	}
	return names[t-1]
}

func (t QueueDirection) Val() int8 {
	return int8(t)
}

type ActionType int8

const (
	Insert ActionType = (1 + iota)
	Update
	Delete
)

func (t ActionType) String() string {
	names := [...]string{
		"insert DB operation",
		"update DB operation",
		"delete DB operation",
	}
	if t < Insert || t > Delete {
		return "unknown DB operation"
	}
	return names[t-1]
}

func (t ActionType) Val() int8 {
	return int8(t)
}
