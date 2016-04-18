package app

import (
    "net"
    "github.com/jinzhu/gorm"
    "github.com/docker/distribution/registry/client"

    "github.com/spacexnice/ctlplane/pro/control"

    "github.com/gin-gonic/gin"
    "github.com/docker/distribution/context"
    "github.com/golang/glog"
    "github.com/spacexnice/ctlplane/pro/web"
    "github.com/spacexnice/ctlplane/pro/util"
    "os"
)

//var REPO_SERVER = "http://61.160.36.122:8080/"
var ENV_REPO_SERVER = "REPO_SERVER"
//var REPO_SERVER = "http://192.168.139.128:5000/"
var REPO_SERVER = "http://192.168.139.131:5000/"

type Config struct {
    Port              int
    Address           net.IP

    DataPath          string
    RegURL            string
    EnableProfiling   bool
    BindPodsQPS       float32
    BindPodsBurst     int
}

type PlaneServer struct {
    Cnf        * Config
    Backend    client.Registry
    // interface to access sqlite db
    DB         * gorm.DB
    Sync       * control.SyncController
    eng        * gin.Engine
    Handler    * web.WebHandler
}

func NewPlaneServer() *PlaneServer{
    cnf := createConfig()
    db  := util.OpenInit()
    r,err := client.NewRegistry(context.Background(),cnf.RegURL,nil)
    if err != nil{
        glog.Errorf("create Registry error while NewPlaneServer ! [%v]",err)
    }
    return &PlaneServer{
        DB:      db,
        Cnf:     cnf,
        Backend: r,
        Sync:    control.NewSyncContoller(cnf.RegURL,db),
        eng:     gin.Default(),
        Handler: web.NewWebHandler(db,cnf.RegURL),
    }
}



func createConfig()* Config{
    rs := os.Getenv(ENV_REPO_SERVER)
    if rs == ""{
        rs = REPO_SERVER
    }
    return &Config{
        RegURL: rs,
        Port:   8080,
        Address:net.IPv4(0,0,0,0),
    }
}

func (s * PlaneServer) AddFlags(){


}
func (s * PlaneServer) route(){
    r := s.eng
    r.Static("/js", "js")
    r.Static("/css", "css")
    r.Static("/fonts", "fonts")
    r.LoadHTMLGlob("pages/*.html")

    r.GET("/detail", s.Handler.GetTag)
    r.POST("/detail", s.Handler.PutTag)
    r.GET("/",s.Handler.Index)
    r.GET("/repository",s.Handler.Index)
    r.GET("/tags/delete", s.Handler.DeleteTag)
    //	r.GET("/namespaces/:ns", listOthersInNamespace)
    r.GET("/help", s.Handler.Help)
    r.GET("/callback/event",s.Handler.Event)
    r.POST("/callback/event",s.Handler.Event)
}


func (s * PlaneServer) Run() error{


    s.Sync.Start()

    s.route()

    s.eng.Run(":8080")

    select {
    }
    return nil
}