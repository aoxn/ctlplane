package web

import (
    "github.com/gin-gonic/gin"
    "github.com/docker/distribution/notifications"
    "fmt"
    "io/ioutil"
)


func (h *WebHandler) Event(c * gin.Context){
    var ev notifications.Envelope
    if c.BindJSON(&ev) == nil{

    }

    exx,err := ioutil.ReadAll(c.Request.Body)
    if err != nil{
        fmt.Println(err.Error()+"xxxxxxxxxxxxxxxx")
    }
    fmt.Println("OK: == "+string(exx))
    fmt.Printf("%+v",ev)
}
