package main

import (
    "fmt"
    "time"
    "os"
"github.com/shirou/gopsutil/process"
)

var ps * process.Process

func main() {
    //t,e := time.Parse("2006-01-02 15:04:05.999999 +0000 UTC","2016-04-29 03:59:28.3553002 +0000 UTC")
    //fmt.Printf("xxx:%v,\nyyy:%v\n",t,e)
    //fmt.Println("YYY",t.Format("2006-01-02 15:04:05.999999999 +0100 MST"))

    Bdate := "2014-06-24 02:30"//时间字符串

    t, err := time.ParseInLocation("2006-01-02 15:04", Bdate, time.UTC)//t被转为本地时间的time.Time
    fmt.Printf("LOCAL:%s\n",time.Unix(t.Unix(),0 ))
    z,off := t.Zone()
    fmt.Printf("Zone:%s, offset:%v\n",z,off)
    //t,err = time.Parse("2006-01-02 15:04", Bdate)//t被转为UTC时间的time.Time
    fmt.Printf("LOCAL:%s\n",time.Unix(t.Unix() + int64(off),0))

    fmt.Printf("xxx:%v,\nyyy:%v\n",t.Format(time.RFC3339),err)
}

func mem(n int){
    if ps == nil{
        p,err := process.NewProcess(int32(os.Getpid()))
        if err != nil{
            panic(err)
        }
        ps = p
    }
    mem , err := ps.MemoryInfo()
    if err != nil{
        panic(err)
    }
    fmt.Printf("%+v\n",mem)
    //fmt.Printf("%d. VMS: %d MB, RSS: %d MB\n",n,mem.VMS>>20,mem.RSS>>20)
}
