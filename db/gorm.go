package db

import (
	"alexchomiak/go-docker-api/types"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Gorm *gorm.DB

func init() {
	Open()
	Gorm.Table("compose_stacks").AutoMigrate(&types.ComposeStack{})
}

func Open() error {
	sqllite_dir := os.Getenv("SQLLITE_PATH")
	if sqllite_dir == "" {
		sqllite_dir = "./server.db"
	}

	var err error
	Gorm, err = gorm.Open(sqlite.Open(sqllite_dir), &gorm.Config{})
	if err != nil {
		return err
	}
	return nil
}

func Close() error {
	db, err := Gorm.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
