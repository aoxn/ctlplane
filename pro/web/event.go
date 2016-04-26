package web

import (
    "github.com/gin-gonic/gin"
    "github.com/docker/distribution/notifications"
    "strings"
    "github.com/Azure/azure-sdk-for-go/core/http"
    "github.com/spacexnice/ctlplane/pro/api"
    "github.com/golang/glog"
    //"encoding/json"
)

const (
    USER_AGENT = "docker"

    USER_MEDIATYPE = "application/vnd.docker.distribution.manifest.v2+json"
)

func (h *WebHandler) Event(c * gin.Context){
    var ev notifications.Envelope
    if err := c.BindJSON(&ev);err != nil{
        h.Retry(c,err)
        return
    }
    //x,_ := json.MarshalIndent(ev, "", "    ")
    //glog.Info(string(x))
    for _,e := range ev.Events{
        glog.Infof("%s   %s   %s   %s    %s    %s    %s   %s",e.Timestamp,e.ID,e.Action,e.Request.UserAgent,e.Target.MediaType,e.Target.Repository,e.Target.Tag,e.Request.Method)

        if strings.Index(e.Request.UserAgent,USER_AGENT) == -1{
            glog.Infoln("Only statistic DOCKER operation")
            c.JSON(http.StatusOK, gin.H{"status": "OK"})
            return
        }
        if strings.Index(e.Target.MediaType,USER_MEDIATYPE) == -1 {
            glog.Infof("Only statistic mediatype=[%s] operation",USER_MEDIATYPE)
            c.JSON(http.StatusOK, gin.H{"status": "OK"})
            return
        }
        if strings.Index(e.Action,"pull") != -1 {
            //pull action , add 1
            r,t,e := h.getRepoTag(e.Target.Repository,e.Target.Tag)
            if e != nil {
                h.Retry(c,e)
                return
            }
            // set pull cnt
            r.TPullCount += 1
            t.PullCount  += 1
            if err := h.DB.Model(&r).Update(api.Repository{TPullCount:r.TPullCount}).Error;err != nil{
                h.Retry(c,err)
                return
            }

            if err := h.DB.Model(&t).Update(api.Tag{PullCount:t.PullCount}).Error; err != nil{
                h.Retry(c,err)
                return
            }
            glog.Infoln("PULL DETECTED !  update pullcount[totalpull:%d, tagpull:%d]",r.TPullCount,t.PullCount)
            c.JSON(http.StatusOK, gin.H{"status": "OK"})
            return
        }
        if strings.Index(e.Action,"push") != -1 {

            //push action , add 1, record push time
            _,t,er := h.getRepoTag(e.Target.Repository,e.Target.Tag)
            if er != nil {
                h.Retry(c,er)
                return
            }
            if err := h.DB.Model(&t).Update(api.Tag{PushTime:e.Timestamp}).Error; err != nil{
                h.Retry(c,err)
                return
            }
            glog.Infoln("PUSH DETECTED !  update timestamp[%s]",e.Timestamp)
            c.JSON(http.StatusOK, gin.H{"status": "OK"})
            return
        }
    }
}

func (h * WebHandler) getRepoTag(rs , ts string) (*api.Repository,*api.Tag,error){
    r := api.Repository{RepoName:rs}
    if err :=h.DB.First(&r).Error;err != nil{
        return nil,nil,err
    }
    t :=  api.Tag{Name:ts,RepositoryID:r.ID}
    if err := h.DB.Where(&t).Find(&t).Error ;err != nil{
        return nil,nil,err
    }
    return &r,&t,nil
}

func (h * WebHandler) Ok(c * gin.Context){
    c.JSON(http.StatusOK, gin.H{"status": "OK"})
}
func (h * WebHandler) Retry(c * gin.Context,e error){
    // registry will continue retry send this event as long as we do not return statusok
    c.JSON(http.StatusInternalServerError, gin.H{"status": e.Error()})
}