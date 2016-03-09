package util

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
//	"k8s.io/kubernetes/pkg/api"
//	"k8s.io/kubernetes/pkg/api/resource"
//	"k8s.io/kubernetes/pkg/util/sets"

//	api_uv "k8s.io/kubernetes/pkg/api/unversioned"
	"github.com/jinzhu/gorm"
	"os"
	"github.com/spacexnice/ctlplane/pkg/api"
	"github.com/spacexnice/ctlplane/pkg/page"

	"time"
)

var DEFAULT_DB = "./gorm.db"

func InitDB() bool{
	if Exist(DEFAULT_DB){
		fmt.Println("DEFAULT_DB[./gorm.db] Loaded! ")
		return true
	}
	db, err := gorm.Open("sqlite3", DEFAULT_DB)
	if err != nil {
		panic(err)
		return false
	}
	db.LogMode(true)
	db.CreateTable(&api.Repository{})
	db.CreateTable(&api.Tag{})
	return true

}
func InitSync(){
	go func() {
		db, _ := gorm.Open("sqlite3", DEFAULT_DB)
		ch := make(chan bool, 0)
		for {
			select {
			case <-time.After(time.Duration(10*time.Second)):
				Run(&db)
			case <-ch:
				fmt.Printf("stopped ,should never happen")
			}
		}
	}()
}

func Run(db *gorm.DB){
	rep,err := page.ListAll()
	if err != nil {
		fmt.Errorf("Syncer Error list all repository,[%s] ",err.Error())
		return
	}
	dbRep := api.Repos{}
	if err := db.Find(&dbRep).Error;err != nil {
		fmt.Errorf("Syncer Error Find all repository in Sqlit3 database, [%s]",err)
		return
	}
	for _,it := range(dbRep){
		var tag api.Tags
		db.Model(it).Related(&tag)
		it.Tags = tag
	}
	var deleted,added api.Repos
	for _,it := range(rep){
		find := false
		for _,dit := range(dbRep){
			if it.RepoName==dit.RepoName {
				find = true
				break
			}
		}
		if !find{
			added = append(added,it)
		}
	}
	for _,dit := range(dbRep){
		find := false
		for _,it := range(rep){
			if it.RepoName==dit.RepoName {
				find = true
				break
			}
		}
		if !find{
			deleted = append(deleted,dit)
		}
	}
	fmt.Printf("added   %+v\ndeleted %+v\n",added,deleted)
	for _,it := range(deleted){
		db.Unscoped().Delete(&it)
	}
	for _,it := range(added){
		db.Create(&it)
	}
}


// 检查文件或目录是否存在
// 如果由 filename 指定的文件或目录存在则返回 true，否则返回 false
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}