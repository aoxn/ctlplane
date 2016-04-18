package control

import (
    "time"
    "github.com/docker/distribution/registry/client"
    "github.com/jinzhu/gorm"
    "github.com/docker/distribution/context"
    "github.com/docker/distribution/reference"
    "fmt"
    "github.com/golang/glog"
    "strings"
    "github.com/spacexnice/ctlplane/pro/util"
    "io"
    _ "github.com/docker/distribution/manifest/schema2"
    "github.com/spacexnice/ctlplane/pro/api"
)

var SYNC_PERIOD = 60 * time.Second

type SyncController struct {
    BaseUrl    string
    Backend    client.Registry
    // interface to access sqlite db
    DB         *gorm.DB
    syncStop   chan struct{}
    syncPeriod time.Duration
}

func NewSyncContoller(url string, db *gorm.DB) *SyncController {
    b, err := client.NewRegistry(context.Background(), url, nil)
    if err != nil {
        glog.Errorf("creat Registry error ! [%v]", err)
    }
    return &SyncController{
        BaseUrl:    url,
        Backend:    b,
        DB:         db,
        syncPeriod: SYNC_PERIOD,
        syncStop:   make(chan struct{}),
    }
}

func (c *SyncController) Start() {
    go util.Until(func() {

        r, err := c.RegRepositories()
        if err != nil {
            glog.Errorf("error occured while get images info from registry, [%v]\n", err)
            return
        }
        d, err := c.DbRepositories()
        if err != nil {
            glog.Errorf("error occured while get images info from database, [%v]\n", err)
            return
        }
        for idx, _ := range r {
            //fmt.Printf("%+v,, %+v\n",idx,r[idx])
            it := r[idx]
            if m, e := d[it.RepoName]; e {
                c.Update(&m, &it)
            }else {
                //not exist in database , create a new one
                c.DB.Create(&it)
            }
        }

        for _, it := range d {
            if m, e := r[it.RepoName]; !e {
                //not exist in registry, we delete it from database
                c.DB.Unscoped().Delete(&m)
            }
        }
        return
    }, c.syncPeriod, c.syncStop)
}

func (c *SyncController) Update(dbRep, regRep *api.Repository) {
    var found bool = false
    for _, dt := range (dbRep.Tags) {
        found = false
        for _, rt := range (regRep.Tags) {
            if dt.Name == rt.Name {
                found = true
                break
            }
        }
        if !found {
            //fmt.Printf("DELETE UpdateTAG: %+v\n",dt)
            if err := c.DB.Unscoped().Delete(&dt).Error; err != nil {
                fmt.Println("DELETE Tag ERROR: %s", err.Error())
            }
        }
    }
    for _, dt := range (regRep.Tags) {
        found = false
        for _, rt := range (dbRep.Tags) {
            if dt.Name == rt.Name {
                found = true
                break
            }
        }
        if !found {
            dt.RepositoryID = dbRep.ID
            //fmt.Printf("ADDED UpdateTAG: %+v\n",dt)
            if err := c.DB.Create(&dt).Error; err != nil {
                fmt.Println("Create Tag ERROR: %s", err)
            }
        }
    }
}

func (c *SyncController) RegRepositories() (map[string]api.Repository, error) {
    entry := make([]string, 1000, 30000)
    repos := make(map[string]api.Repository)
    cnt, err := c.Backend.Repositories(context.Background(), entry, "")
    if err != nil&&err != io.EOF {
        return nil, err
    }
    //fmt.Printf("cnt::  %v\n",cnt)
    for i := 0; i < cnt; i++ {
        r := api.Repository{RepoName:entry[i]}
        ctx := context.Background()
        repo, _ := reference.ParseNamed(entry[i])
        if rp, err := client.NewRepository(ctx, repo, c.BaseUrl, nil); err == nil {
            //fmt.Printf("REPOSITORY: %+v\n", rp)
            ts := rp.Tags(ctx)

            all, err := ts.All(ctx)
            //fmt.Printf("TAGS: %+v\n", all)
            if err == nil {
                for _, t := range all {
                    if des, err := ts.Get(ctx, t); err == nil {
                        //fmt.Printf("Tag ::  %+v\n", des)
                        tag := api.Tag{Name:t, Digest:des.Digest.String()}
                        r.Tags = append(r.Tags, tag)
                    }
                }
            }
            parseGroup(&r)
            repos[r.RepoName] = r
        }

    }
    return repos, nil
}

func (c *SyncController) DbRepositories() (map[string]api.Repository, error) {
    var dbRep []api.Repository
    if err := c.DB.Find(&dbRep).Error; err != nil {
        glog.Errorf("Syncer Error Find all repository in Sqlit3 database, [%s]", err)
        return nil, err
    }
    for idx, _ := range (dbRep) {
        var tag []api.Tag
        if err := c.DB.Model(&dbRep[idx]).Related(&tag).Error; err != nil {
            glog.Errorf("Error: related %+v %+v\n", err, tag)
        }
        dbRep[idx].Tags = tag
    }
    rt := make(map[string]api.Repository)
    for _, it := range dbRep {
        rt[it.RepoName] = it
    }
    return rt, nil
}

func parseGroup(r *api.Repository) {
    n := strings.Split(r.RepoName, "/")
    if len(n) <= 1 {
        r.Group = "default"
    }else {
        r.Group = n[0]
    }
    //fmt.Printf("Group: %s\n", r.Group)
}
