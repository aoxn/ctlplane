package util

import (
    "github.com/jinzhu/gorm"
    "github.com/golang/glog"
    "github.com/spacexnice/ctlplane/pro/api"
    "os"
    _ "github.com/mattn/go-sqlite3"
    "fmt"
)

var DEFAULT_DB = "gorm.db"

func open(path string) *gorm.DB {
    db, err := gorm.Open("sqlite3", path)
    if err != nil {
        panic(err)
        return nil
    }
    return db
}
func OpenInit(path string) *gorm.DB {
    var db *gorm.DB
    fdb := fmt.Sprintf("%s/%s",path,DEFAULT_DB)
    if Exist(fdb) {
        glog.Info("DATABASE INIT: database [%s] exist,load tables... ",fdb)
        db = open(fdb)
        return db
    }
    if db = open(fdb); db == nil {
        return nil
    }
    glog.Info("DATABASE INIT: database [%s] dose not exist,create database and tables... ",fdb)
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