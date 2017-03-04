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
    "flag"
)

//var REPO_SERVER = "http://61.160.36.122:8080/"
var (
    ENV_REPO_SERVER = "REPO_SERVER"
    ENV_DB_DATA_PATH= "DB_DATA_PATH"
)
//docker run -d -p 5000:5000 --name registry registry:2
//var REPO_SERVER = "http://192.168.139.128:5000/"
var (
    REPO_SERVER = "http://hubregistry/"
)

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
    db  := util.OpenInit(cnf.DataPath)
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
    dpath := os.Getenv(ENV_DB_DATA_PATH)
    if dpath == ""{
        p,_ := os.Getwd()
        dpath = p
    }
    return &Config{
        DataPath:   dpath,
        RegURL:     rs,
        Port:       8080,
        Address:    net.IPv4(0,0,0,0),
    }
}

func (s * PlaneServer) AddFlags(){
    flag.Set("logtostderr", "true")
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

    r.GET("/test",s.Handler.PTest)

    //// Authorization group
    //// authorized := r.Group("/", AuthRequired())
    //// exactly the same as:
    //authorized := r.Group("/")
    //// per group middleware! in this case we use the custom created
    //// AuthRequired() middleware just in the "authorized" group.
    //authorized.Use(AuthRequired())
    //{
    //    authorized.POST("/login", loginEndpoint)
    //    authorized.POST("/submit", submitEndpoint)
    //    authorized.POST("/read", readEndpoint)
    //
    //    // nested group
    //    testing := authorized.Group("testing")
    //    testing.GET("/analytics", analyticsEndpoint)
    //}
}


func (s * PlaneServer) Run(){
    s.Sync.Start()
    s.route()
    s.eng.Run(":8080")
}
