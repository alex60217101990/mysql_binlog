package mysql_binlog

type GameLogo struct {
	ID          int64  `json:"-" gorm:"primary_key;column:id;" sql:"type:int unsigned auto_increment"`
	GameID      int64  `json:"game_id" gorm:"column:game_id;type:bigint;"`
	GameLogoURL string `json:"game_logo" gorm:"type:varchar(500);column:game_logo;"`
}

func (GameLogo) TableName() string {
	return "games_logos"
}

func (GameLogo) FromBinLog(fields ...interface{}) interface{} {
	g := GameLogo{}
	g.ID = int64(fields[0].(int32))
	g.GameID = fields[1].(int64)
	g.GameLogoURL = fields[2].(string)
	return g
}
