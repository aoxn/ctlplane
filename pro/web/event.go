package web

import (
    "github.com/gin-gonic/gin"
    "github.com/docker/distribution/notifications"
    "strings"
    "github.com/spacexnice/ctlplane/pro/api"
    "github.com/golang/glog"
    //"encoding/json"
    "net/http"
    "fmt"
    "errors"
    "github.com/docker/distribution/manifest/schema2"
    "github.com/docker/distribution/manifest/schema1"
)

const (
    USER_AGENT = "docker"
)

func (h *WebHandler) Event(c * gin.Context){
    var ev notifications.Envelope
    if err := c.BindJSON(&ev);err != nil{
        glog.Errorf("Wrong Type,receive no json object [%s]",err.Error())
        h.Retry(c,err)
        return
    }
    //x,_ := json.MarshalIndent(ev, "", "    ")
    //glog.Info(string(x))
    for _,e := range ev.Events{
        glog.Infof("%s   %s   %s   %s    %s    %s    %s   %s",e.Timestamp.Format("2006-01-02 15:04:05"),e.ID,e.Action,e.Request.UserAgent,e.Target.MediaType,e.Target.Repository,e.Target.Tag,e.Request.Method)

        if h.skip(e){
            h.Ok(c)
            return
        }

        if err := h.doPull(e); err != nil{
            h.Retry(c,err)
            return
        }
        if err := h.doPush(e); err != nil{
            h.Retry(c,err)
            return
        }
    }
    h.Ok(c)
}




func (h * WebHandler) skip(e notifications.Event) bool{
    if strings.Index(e.Request.UserAgent,USER_AGENT) == -1{
        glog.Infoln("Only statistic DOCKER operation")

        return true
    }
    return false
}
func (h * WebHandler) doPull(e notifications.Event) error{
    if strings.Index(e.Action,"pull") == -1 {
        //process ok if not pull
        return nil
    }
    if strings.Index(e.Target.MediaType,schema1.MediaTypeManifest) == -1 &&
    strings.Index(e.Target.MediaType,schema1.MediaTypeSignedManifest) == -1 &&
    strings.Index(e.Target.MediaType,schema2.MediaTypeManifest) == -1 {

        // Only process MediaType of manifests.
        glog.Infof("Only statistic mediatype=[%s][%s][%s] operation",
            schema1.MediaTypeManifest,schema1.MediaTypeSignedManifest,schema2.MediaTypeManifest)

        return nil
    }
    //pull action , add 1
    r,t,er := h.getRepoTag(e.Target.Repository,e.Target.Tag)
    if er != nil {
        return er
    }
    // set pull cnt
    r.TPullCount += 1
    t.PullCount  += 1
    if err := h.DB.Model(&r).Update(api.Repository{TPullCount:r.TPullCount}).Error;err != nil{
        return err
    }

    if err := h.DB.Model(&t).Update(api.Tag{PullCount:t.PullCount}).Error; err != nil{
        return err
    }
    glog.Infoln("PULL DETECTED !  update pullcount[totalpull:%d, tagpull:%d]",r.TPullCount,t.PullCount)
    return nil
}

func (h * WebHandler) doPush(e notifications.Event) error{
    if strings.Index(e.Action,"push") == -1 {
        //process ok if not push
        return nil
    }
    //push action , add 1, record push time
    _,t,er := h.getRepoTag(e.Target.Repository,e.Target.Tag)
    if er != nil {
        glog.Errorf("PUSH: error get tag[%s][%s]",e.Target.Repository,e.Target.Tag)
        return er
    }
    if err := h.DB.Model(&t).Update(api.Tag{PushTime:e.Timestamp}).Error; err != nil{
        glog.Errorf("PUSH: error update pushtime.[%s][%s]",e.Target.Repository,e.Target.Tag)
        return err
    }
    glog.Infoln("PUSH DETECTED !  update timestamp[%s]",e.Timestamp.Format("2006-01-02 15:04:05"))
    return nil
}

func (h * WebHandler) getRepoTag(rs , ts string) (*api.Repository,*api.Tag,error){
    r := api.Repository{RepoName:rs}
    if err :=h.DB.Where(&r).First(&r).Error;err != nil{
        glog.Errorf("PUSH: error select Repo[%s][%s]",rs,ts)
        return nil,nil,err
    }
    t :=  api.Tag{Name:ts,RepositoryID:r.ID}
    if err := h.DB.Where(&t).First(&t).Error ;err != nil{
        glog.Errorf("PUSH: error select tag[%s][%s]",rs,ts)
        return nil,nil,err
    }
    if t.ID == 0{
        // Tag not found , maybe first put, just let notification retry
        return nil,nil,errors.New(fmt.Sprintf("Event CallBack: Tag not found , maybe first push, just let notification retry![%s][%s]",r.RepoName,ts))
    }
    return &r,&t,nil
}

func (h * WebHandler) Ok(c * gin.Context){
    c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
func (h * WebHandler) Retry(c * gin.Context,e error){
    // registry will continue retry send this event as long as we do not return statusok
    glog.Errorf("Notifcation error,need retry! [%s]",e.Error())
    c.JSON(http.StatusInternalServerError, gin.H{"status": e.Error()})
}