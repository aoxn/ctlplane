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
	"strings"
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

func ImageDeletor(){

}
func Run(db *gorm.DB){
	rep,err := page.ListAll()
	if err != nil {
		fmt.Errorf("Syncer Error list all repository,[%s] ",err.Error())
		return
	}
	var dbRep []api.Repository
	if err := db.Find(&dbRep).Error;err != nil {
		fmt.Errorf("Syncer Error Find all repository in Sqlit3 database, [%s]",err)
		return
	}
	for idx,_ := range(dbRep){
		var tag []api.Tag
		if err:=db.Model(&dbRep[idx]).Related(&tag).Error;err != nil{
			fmt.Printf("Error: related %+v %+v\n",err,tag)
		}
		dbRep[idx].Tags = tag
	}
	for idx,_ := range(rep){
		find := false
		for _,dit := range(dbRep){
			if rep[idx].RepoName==dit.RepoName {
				find = true
				break
			}
		}
		if !find{
			parseGroup(&rep[idx])
			db.Create(&rep[idx])
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
			db.Unscoped().Delete(&dit)
		}
	}

	for _,dit := range(dbRep){
		for _,it := range(rep){
			if it.RepoName==dit.RepoName {
				updateTag(&dit,&it,db)
				break
			}
		}
	}

	fmt.Printf("DB DATA: => \n\tdbRep:   %+v\n\trep: %+v\n",dbRep,rep)
}

func updateTag(dbRep,regRep *api.Repository,db *gorm.DB){
	var found bool = false
	for _,dt := range(dbRep.Tags){
		found = false
		for _,rt := range(regRep.Tags){
			if dt.Name == rt.Name{
				found = true
				break
			}
		}
		if !found{
			fmt.Printf("DELETE UpdateTAG: %+v\n",dt)
			if err:=db.Unscoped().Delete(&dt).Error;err != nil{
				fmt.Println("DELETE Tag ERROR: %s",err.Error())
			}
		}
	}
	for _,dt := range(regRep.Tags){
		found = false
		for _,rt := range(dbRep.Tags){
			if dt.Name == rt.Name{
				found = true
				break
			}
		}
		if !found{
			dt.RepositoryID = dbRep.ID
			fmt.Printf("ADDED UpdateTAG: %+v\n",dt)
			if err:=db.Create(&dt).Error ;err!=nil{
				fmt.Println("Create Tag ERROR: %s",err)
			}
		}
	}
}

func parseGroup(r *api.Repository){
	n := strings.Split(r.RepoName,"/")
	if len(n)<=1{
		r.Group = "default"
	}else{
		r.Group = n[0]
	}
	fmt.Printf("Group: %s\n",r.Group)
}

// 检查文件或目录是否存在
// 如果由 filename 指定的文件或目录存在则返回 true，否则返回 false
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}