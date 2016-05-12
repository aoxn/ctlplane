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
    "github.com/docker/distribution/manifest/schema1"
    "github.com/golang/glog"
)



func (h *WebHandler)  PTest(c *gin.Context){
    h.imageLayers("admin/bob","v2")
}

func (h * WebHandler) imageLayers(vrepo,vtag string)([]api.Layer,error){
    ctx := context.Background()
    repo, err := reference.ParseNamed(vrepo)
    if err != nil {
        return nil,err
    }
    rp, err := client.NewRepository(ctx, repo, h.RegURL, nil)
    if err != nil {
        return nil,err
    }

    descriptor, err := rp.Tags(ctx).Get(ctx,vtag)
    if err != nil{
        return nil,err
    }

    msvc,err := rp.Manifests(ctx)
    if err != nil{
        return nil,err
    }

    manifest,err := msvc.Get(ctx,descriptor.Digest)
    if err != nil{
        return nil,err
    }

    switch manifest.(type) {
    case *schema1.SignedManifest:
        return h.v1Layer(manifest.(*schema1.SignedManifest)),nil
    case *schema2.DeserializedManifest:
        cfg,smanifest := &image.Image{},manifest.(*schema2.DeserializedManifest)
        bconfig,err   := rp.Blobs(ctx).Get(ctx,smanifest.Config.Digest)
        if err != nil{
            return nil,err
        }

        e:=json.Unmarshal(bconfig,cfg)
        if e != nil{
            fmt.Println("Wrong Type of Image!",e.Error())
        }
        return h.v2layer(cfg,smanifest),nil
    }
    return nil,nil
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

func (h *WebHandler) v2layer(img *image.Image,mani *schema2.DeserializedManifest) []api.Layer{
    var layer []api.Layer
    for k,v := range img.History{
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
        lsize := "-1"
        for _,vm := range mani.References(){
            if mani.Layers[k].Digest == vm.Digest{
                lsize = fmt.Sprintf("%d",vm.Size/1024)
                break
            }
        }

        //fmt.Println(v.CreatedBy,CMD_ADD,"HHH",strings.Index(v.CreatedBy,CMD_ADD),color)
        layer = append(layer,api.Layer{
            Created:        v.Created,
            CreatedBy:      strings.Replace(v.CreatedBy,CMD,"",-1),
            Author:         v.Author,
            Color:          color,
            Size:           lsize,
        })
    }
    return layer
}


func (h *WebHandler) v1Layer(mani *schema1.SignedManifest)[]api.Layer{
    var layers []api.Layer
    for k,v := range mani.History{
        v1image := image.V1Image{}
        err := json.Unmarshal([]byte(v.V1Compatibility),&v1image)
        if err != nil {
            glog.Errorf("Unmarshal V1Compatibility Error [%s]\n",err.Error())
            continue
        }
        color,scmd := "",""
        for _,s := range v1image.ContainerConfig.Cmd{
            scmd += fmt.Sprintf("%s ",s)
        }
        if strings.Index(scmd,CMD_ADD) != -1 {
            color = COLOR_INFO
        }
        if strings.Index(scmd,CMD_CMD) != -1 {
            color = COLOR_WARNING
        }
        if strings.Index(scmd,CMD_COPY) != -1 {
            color = COLOR_SUCCESS
        }
        if strings.Index(scmd,CMD_ENV) != -1 {
            color = COLOR_DANGER
        }

        lsize := "-1"
        for _,vm := range mani.References(){
            if mani.FSLayers[k].BlobSum == vm.Digest{
                lsize = fmt.Sprintf("%d",vm.Size/1024)
                break
            }
        }

        layers = append(layers,api.Layer{
            Created:    v1image.Created,
            CreatedBy:  strings.Replace(scmd,CMD,"",-1),
            Author:     v1image.Author,
            Color:      color,
            Size:       lsize,
        })
    }
    return layers
}
