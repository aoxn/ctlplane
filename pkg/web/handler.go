package web
import (
    "github.com/gin-gonic/gin"
    "fmt"
    "net/http"
    "strconv"
    "github.com/spacexnice/ctlplane/pkg/api"
    "github.com/spacexnice/ctlplane/pkg/page"
    "github.com/jinzhu/gorm"
    "github.com/spacexnice/ctlplane/pkg/util"
)

var (
    db gorm.DB
)

func init(){
    util.InitDB()
    var err error
    db, err = gorm.Open("sqlite3", util.DEFAULT_DB)
    db.LogMode(true)
    if err != nil {
        panic(err)
    }
}

func getPager(idx string) *page.Pager{
    pg := &page.Pager{Size:page.PAGE_SIZE}
    if p,err := strconv.Atoi(idx);err ==nil{
        if p > 1{
            pg.Index = p
            pg.Previous = pg.Index - 1
            pg.Next = pg.Index +1
            return pg
        }
    }
    pg.Index = 1
    pg.Previous = 1
    pg.Next  = 2
    return pg
}

func Index(c *gin.Context){
//    fmt.Printf("%+v\n",c.Request)

    var result []api.Repository

    idx,_ := c.GetQuery("index")
    pg:= getPager(idx)

    sea,suc := c.GetQuery("search")
    if suc{
        pg.Search = sea
        if err := db.Where("repo_name like ?",fmt.Sprintf("%%%s%%",sea)).
                  Limit(pg.Size).
                  Offset(pg.Size*(pg.Index - 1)).
                  Find(&result).Error;err != nil{
            errorPage(c,fmt.Sprintf("Get Repository Error,[%s]",err.Error()))
            return
        }
    }else{
        if err := db.Limit(pg.Size).
                  Offset(pg.Size*(pg.Index - 1)).
                  Find(&result).Error;err != nil{
            errorPage(c,fmt.Sprintf("Get Repository Error,[%s]",err.Error()))
            return
        }
    }
    pg.Data = result
    c.HTML(http.StatusOK, "index",
        gin.H{
            "title":     "REPOSITORY",
            "has":true,
            "page":pg,
    })
    return
}


func GetTag(c *gin.Context) {
    repoName,suc := c.GetQuery("it")
    if !suc{
        errorPage(c,"error parameter repository name needed by it=?")
        return
    }
    repo,err := getRepository(repoName)
    if err != nil{
        errorPage(c,fmt.Sprintf("error get repository name needed by it=?,[%s]",err.Error()))
        return
    }
    p,_ := c.GetQuery("tag")
    tag := getSelectedTag(repo.Tags,p)
    if tag == nil{
        errorPage(c,"No Tags Found!")
        return
    }
    fmt.Printf("TAGS: %+v \n", tag)
    c.HTML(http.StatusOK, "tags",
        gin.H{
            "title":     "REPOSITORY",
            "repo" : repo,
            "has": len(repo.Tags)>0,
            "currtag":tag,
        })
    return
}

func errorPage(c *gin.Context,error string){
    c.HTML(http.StatusOK, "tags",
        gin.H{
            "title": "Error",
            "repo" : "",
            "has": false,
            "currtag": "",
            "errorinfo": error,
        })
}

func PutTag(c *gin.Context) {
    stag,suc := c.GetQuery("tag")
    if !suc{
        errorPage(c,"error parameter tag name needed by it=?")
        return
    }
    txt,suc := c.GetPostForm("txtbody")
    if !suc {
        errorPage(c,"error parameter txtbody name needed by it=?")
        return
    }
    repoName,suc := c.GetQuery("it")
    if !suc{
        errorPage(c,fmt.Sprintf("error parameter repository name needed by it"))
        return
    }
    repo,err := getRepository(repoName)
    if err != nil{
        errorPage(c,fmt.Sprintf("error get Repository by it=?[%s]",err.Error()))
        return
    }

    tag := getSelectedTag(repo.Tags,stag)
    if tag == nil{
        errorPage(c,fmt.Sprintf("No curresponed Tag found by it=?[%s]",tag))
        return
    }
    tag.Description = txt
    if err := db.Save(tag).Error; err !=nil{
        errorPage(c,fmt.Sprintf("error save Description by it=?[%s]",err.Error()))
        return
    }
    c.HTML(http.StatusOK, "tags",
        gin.H{
            "title":     "REPOSITORY",
            "repo" : repo,
            "has": len(repo.Tags)>0,
            "currtag":tag,
        })
    return
}

func getRepository(repoName string) (*api.Repository,error){
    tags := api.Tags{}
    repo := api.Repository{RepoName:repoName}
    if err := db.Where("repo_name = ?",repoName).First(&repo).Error;err!=nil{
        fmt.Errorf("Find repository [%s] Error,[%+v]",repoName,err)
        return nil,err
    }
    if err := db.Model(&repo).Related(&tags).Error ; err != nil{
        fmt.Errorf("Find Tags [%s] Error,[%+v]",repoName,err)
        return nil,err
    }
    repo.Tags = tags
    return &repo,nil
}
func getSelectedTag(tags api.Tags,tag string) *api.Tag{
    var ctag api.Tag
    if tag == ""{
        //return default first one
        if len(tags) > 0{
            return &tags[0]
        }
        return nil
    }
    // Search for right tag
    var found bool
    for _,i := range(tags){
        if tag == i.Name{
            ctag = i
            found = true
            break
        }
    }
    fmt.Println(found)
    if found {
        return &ctag
    }else {
        return nil
    }
}


func Help(c *gin.Context) {
    c.HTML(http.StatusOK, "help", gin.H{
        "title": "Sigma Help",
    })
}
