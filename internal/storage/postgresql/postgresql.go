package postgresql

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"time"
)

type Postgres struct {
	Db *gorm.DB
}

type Command struct {
	Id      int    `gorm:"primary_key;auto_increment"`
	Command string `gorm:"type:text"`
}

type Outputs struct {
	CommandId int       `gorm:"foreignKey:id;references:id"`
	Data      string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func NewPostgresRepository() *Postgres {
	db, err := gorm.Open(postgres.Open(os.Getenv("POSTGRES")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	rawDB, _ := db.DB()
	rawDB.SetMaxOpenConns(128)
	rawDB.SetMaxIdleConns(256)

	if err != nil {
		panic("couldn't connect to database: " + err.Error())
	}
	if err := db.AutoMigrate(&Command{}, &Outputs{}); err != nil {
		panic("can't migrate databases")
	}
	return &Postgres{db}
}
