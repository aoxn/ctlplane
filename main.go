package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
//	"github.com/spacexnice/ctlplane/pkg/web"
	"github.com/spacexnice/ctlplane/pkg/util"
//	"github.com/spacexnice/ctlplane/pkg/api"
//	"fmt"
//	"github.com/spacexnice/ctlplane/pkg/page"
//	"github.com/gin-gonic/gin"
	"github.com/spacexnice/ctlplane/pkg/web"
)


var (
	Db gorm.DB
)

func main() {
	defer glog.Flush()

	flag.Set("logtostderr", "true")
	flag.Parse()

	util.InitDB()

//	util.InitSync()
//	web.GetRepository("bamboo/controller")
	r := gin.Default()
	r.Static("/js", "js")
	r.Static("/css", "css")
	r.Static("/fonts", "fonts")
	r.LoadHTMLGlob("pages/*.html")

	r.GET("/detail", web.GetTag)
	r.POST("/detail", web.PutTag)
	r.GET("/repository",web.Index)
//	r.GET("/namespaces/:ns", listOthersInNamespace)


	r.GET("/help", web.Help)

	r.Run(":8080") // listen and serve on 0.0.0.0:8080
}

