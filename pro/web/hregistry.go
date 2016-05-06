package web

import (
    "github.com/docker/docker/image"
    "github.com/docker/distribution/reference"
    "github.com/docker/distribution/registry/client"
    "github.com/docker/distribution/context"
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/docker/distribution/manifest/schema2"
    "github.com/spacexnice/ctlplane/pro/api"
    "strings"
"encoding/json"
)



func (h *WebHandler)  PTest(c *gin.Context){
    h.image("admin/bob","v2")
}

func (h * WebHandler) image(vrepo,vtag string)(*schema2.DeserializedManifest,*image.Image,error){
    ctx := context.Background()
    repo, err := reference.ParseNamed(vrepo)
    if err != nil {
        return nil,nil,err
    }
    rp, err := client.NewRepository(ctx, repo, h.RegURL, nil)
    if err != nil {
        return nil,nil,err
    }

    descriptor, err := rp.Tags(ctx).Get(ctx,vtag)
    if err != nil{
        return nil,nil,err
    }

    msvc,err := rp.Manifests(ctx)
    if err != nil{
        return nil,nil,err
    }

    mf,err := msvc.Get(ctx,descriptor.Digest)
    if err != nil{
        return nil,nil,err
    }

    smanifest,cfg := mf.(*schema2.DeserializedManifest),&image.Image{}

    bconfig,err   := rp.Blobs(ctx).Get(ctx,smanifest.Config.Digest)
    if err != nil{
        return nil,nil,err
    }

    e:=json.Unmarshal(bconfig,cfg)
    if e != nil{
        fmt.Println("Wrong Type of Image!",e.Error())
    }
    //for k,v := range cfg.History{
    //    fmt.Println(k,":HHHH:",v)
    //}

    return smanifest,cfg,err
}

const (
    COLOR_INFO    = "list-group-item-info"
    COLOR_WARNING = "list-group-item-warning"
    COLOR_SUCCESS = "list-group-item-success"
    COLOR_DANGER  = "list-group-item-danger"

    CMD_ADD = "/bin/sh -c #(nop) ADD"
    CMD_CMD = "/bin/sh -c #(nop) CMD"
    CMD_COPY= "/bin/sh -c #(nop) COPY"
    CMD_MAINTAINER = "/bin/sh -c #(nop) MAINTAINER"
    CMD_ENV = "/bin/sh -c #(nop) ENV"
    CMD     = "/bin/sh -c #(nop) "
)

func (h *WebHandler) layer(img *image.Image) []api.Layer{
    var layer []api.Layer
    for _,v := range img.History{
        color := ""
        if strings.Index(v.CreatedBy,CMD_ADD) != -1 {
            color = COLOR_INFO
        }
        if strings.Index(v.CreatedBy,CMD_CMD) != -1 {
            color = COLOR_WARNING
        }
        if strings.Index(v.CreatedBy,CMD_COPY) != -1 {
            color = COLOR_SUCCESS
        }
        if strings.Index(v.CreatedBy,CMD_ENV) != -1 {
            color = COLOR_DANGER
        }
        //fmt.Println(v.CreatedBy,CMD_ADD,"HHH",strings.Index(v.CreatedBy,CMD_ADD),color)
        layer = append(layer,api.Layer{
            Created:        v.Created,
            CreatedBy:      strings.Replace(v.CreatedBy,CMD,"",-1),
            Author:         v.Author,
            Color:          color,
        })
    }
    return layer
}
