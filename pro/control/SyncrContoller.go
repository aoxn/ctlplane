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
    "flag"
"github.com/docker/distribution"
"github.com/docker/distribution/digest"
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

func init()  {
    flag.Set("logtostderr", "true")
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

    var dbtag map[string]*api.Tag = make(map[string]*api.Tag)
    var rgtag map[string]*api.Tag = make(map[string]*api.Tag)
    for idx,_ := range dbRep.Tags {
        dbtag[fmt.Sprintf("%s/%s",dbRep.RepoName,dbRep.Tags[idx].Name)] = &dbRep.Tags[idx]
    }
    for idx,_ := range regRep.Tags {
        rgtag[fmt.Sprintf("%s/%s",regRep.RepoName,regRep.Tags[idx].Name)] = &regRep.Tags[idx]
    }

    for k,v := range dbtag{
        if _,e := rgtag[k];!e{
            // Found tag not in registry any more,delete it from database
            glog.Infof("UPDATE_TAG: DATABASE DELETE  [%+v]\n",v)
            if err := c.DB.Unscoped().Delete(&v).Error; err != nil {
                glog.Errorln("UPDATE_TAG: DATABASE DELETE ERROR: [%s]", err.Error())
            }
        }
    }

    for k,v := range rgtag{
        if _,e := dbtag[k]; !e{
            v.RepositoryID = dbRep.ID
            v.PushTime = time.Now()
            glog.Infof("UPDATE_TAG: DATABASE ADD [%+v][%s]\n",v,v.PushTime)
            if err := c.DB.Create(&v).Error; err != nil {
                glog.Errorln("UPDATE_TAG: DATABASE ADD ERROR: [%s]", err)
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
        if e := c.getImage(context.Background(),&r);e !=nil{
            return nil,e
        }

        repos[r.RepoName] = r
    }
    return repos, nil
}

func (c *SyncController) getImage(ctx context.Context,r *api.Repository)error{
    repo, err := reference.ParseNamed(r.RepoName)
    if err != nil {
        return err
    }
    rp, err := client.NewRepository(ctx, repo, c.BaseUrl, nil)
    if err != nil {
        return err
    }
    //fmt.Printf("REPOSITORY: %+v\n", rp)
    all, err := rp.Tags(ctx).All(ctx)
    //fmt.Printf("TAGS: %+v\n", all)
    if err != nil{
        return err
    }

    for _, t := range all {
        des, err := rp.Tags(ctx).Get(ctx, t);
        if err != nil {
            return err
        }

        //fmt.Printf("Tag ::  %+v\n", des)
        tag := api.Tag{
            Name:       t,
            Digest:     des.Digest.String(),
            PushTime:   time.Now(),
            Size:       c.getTagSize(rp,ctx,des.Digest,t),
        }
        r.Tags = append(r.Tags, tag)
    }
    parseGroup(r)
    return nil
}

func (c *SyncController) getTagSize(rp distribution.Repository,ctx context.Context,des digest.Digest,tag string)int64{
    var ok int64 = 0
    m,err  := rp.Manifests(ctx)
    if err != nil {
        glog.Errorf("Error while get tagsize[GET LAYERS] [%s]",err.Error())
        return -1
    }
    mi,err := m.Get(ctx,des)
    if err != nil {
        glog.Errorf("Error while get tagsize[GET MANIFEST] [%s]",err.Error())
        return -1
    }
    for _,d := range mi.References(){
        //glog.Info(tag,des,d.Digest,d.Size,d.MediaType)
        ok += d.Size
    }
    return ok/1024
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
