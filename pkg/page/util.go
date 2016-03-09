package page
import (
    "math"
    "fmt"
    "github.com/jinzhu/gorm"
)



type Data []interface{}

type BigPager struct {
    Pre  int
    Curr int
    Next int
    TotalNum  int
    TotalPage int
    Search string
    Data Data
}

func NewBigPager(index int,list func() (Data,error)) (*BigPager , error){
    repo , err:= list()
    if err != nil{
        return nil,err
    }
    totalnum  := len(repo)
    totalpage := int(math.Ceil(float64(totalnum)/PAGE_SIZE))
    pre,next := 0,0
    if index > totalpage{
        index = totalpage
        next = totalpage
    }else{
        next = index + 1
    }
    if index < 0 {
        index = 0
        pre = 0
    }else{
        pre = index -1
    }

    return &BigPager{
        Pre:pre,
        Curr:index,
        Next:next,
        Data:repo,
        TotalNum:totalnum,
        TotalPage:totalpage,
    },nil
}

func (page * BigPager) SetSearch(name string, contain func(item interface{},s string)bool){

    page.Search = name
    var items Data
    fmt.Println(name," Before Search: ",page.Data)
    for _,it := range page.Data{
        if contain(it,page.Search){
            items = append(items,it)
        }
    }
    fmt.Println(name, " After Search: ",items)
    page.Data=items
    page.TotalNum    = len(items)
    page.TotalPage   = int(math.Ceil(float64(page.TotalNum)/PAGE_SIZE))
}

func (page * BigPager) GetPage() *BigPager{
    start := page.Pre * PAGE_SIZE
    end   := page.Curr * PAGE_SIZE
    if start > page.TotalNum{
        start = page.TotalNum-1
    }
    if end > page.TotalNum{
        end = page.TotalNum
    }
    if start < 0{
        start = 0
    }
    if end < 0 {
        end = 0
    }
    var items Data
    for index :=start;index <end;index++{
        items = append(items,page.Data[index])
    }
    page.Data = items
    return page
}

const PAGE_SIZE = 5

type Pager struct {
    Previous int
    Search string
    Index int
    Next int
    Size  int
    Db *gorm.DB
    Data interface{}
}

func NewPager(db *gorm.DB)*Pager{

    return &Pager{
        Index:1,
        Db:db.Limit(PAGE_SIZE),
        Size:PAGE_SIZE,
    }
}

func (p * Pager) WithPageNum(index int) *Pager{
    if index <1 {
        p.Index = 1
    }
    p.Index = index
    p.Db = p.Db.Offset(p.Size*(p.Index - 1))
    return p
}

func (p * Pager) WithSearch(query interface{}, args ...interface{}) *Pager{
    p.Db = p.Db.Where(query,args)
    return p
}

func (p * Pager) List(obj interface{}) *Pager{
    p.Db.Find(obj)
    p.Data = obj
    return p
}

