package main

import (
    "fmt"
    "time"
    "os"
"github.com/shirou/gopsutil/process"
)

var ps * process.Process

func main() {
    fmt.Println(time.Now().Format("2006-01-02 15:04:05"))
    mem(1)

    data:=new([10][1024*1024]byte)
    mem(2)

    //
    for i := range data{
        for x,n := 0,len(data[i]);x<n ;x++{
            data[i][x]=1
        }
        mem(3)
    }
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
