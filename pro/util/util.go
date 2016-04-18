package util

import (
    "github.com/jinzhu/gorm"
    "github.com/golang/glog"
    "github.com/spacexnice/ctlplane/pro/api"
    "os"
    _ "github.com/mattn/go-sqlite3"
)

var DEFAULT_DB = "./gorm.db"

func open() *gorm.DB {
    db, err := gorm.Open("sqlite3", DEFAULT_DB)
    if err != nil {
        panic(err)
        return nil
    }
    return db
}
func OpenInit() *gorm.DB {
    var db *gorm.DB
    if Exist(DEFAULT_DB) {
        glog.Info("DEFAULT_DB[./gorm.db] Loaded! ")
        db = open()
        return db
    }
    if db = open(); db == nil {
        return nil
    }
    db.LogMode(true)
    db.CreateTable(&api.Repository{})
    db.CreateTable(&api.Tag{})
    return db
}

// 检查文件或目录是否存在
// 如果由 filename 指定的文件或目录存在则返回 true，否则返回 false
func Exist(filename string) bool {
    _, err := os.Stat(filename)
    return err == nil || os.IsExist(err)
}