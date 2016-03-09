package idb

import (
	"github.com/jinzhu/gorm"
)

func initDb(){
	db, err := gorm.Open("sqlite3", "/tmp/gorm.db")
}