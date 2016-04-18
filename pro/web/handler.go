package web

import (
    "github.com/gin-gonic/gin"
    "net/http"
    "fmt"
    "github.com/spacexnice/ctlplane/pro/api"
    "github.com/docker/distribution/registry/client"
    "github.com/jinzhu/gorm"
    "strconv"
    "github.com/docker/distribution/context"
"github.com/docker/distribution/digest"
"github.com/docker/distribution/reference"
)

type WebHandler struct {
    RegURL   string
    DB       * gorm.DB
}
type Pager struct {
    Previous int
    Search   string
    Index    int
    Next     int
    Size     int
    Db       *gorm.DB
    Data     interface{}
}

func NewWebHandler(db * gorm.DB,url string) * WebHandler{

    return &WebHandler{
        DB:     db,
        RegURL: url,
    }
}

const (
    QUERY_TAG         = "tag"
    QUERY_REPOSITORY  = "it"
    QUERY_SEARCH      = "search"
    QUERY_INDEX       = "index"

    PAGE_SIZE = 5
)
func (h * WebHandler) Index(c *gin.Context) {
    //    fmt.Printf("%+v\n",c.Request)

    var result []api.Repository

    idx, _ := c.GetQuery(QUERY_INDEX)
    pg := h.getPager(idx)
    sea, suc := c.GetQuery(QUERY_SEARCH)
    if suc {
        pg.Search = sea
        if err := h.DB.Where("repo_name like ?", fmt.Sprintf("%%%s%%", sea)).
        Find(&result).Error; err != nil {
            c.HTML(http.StatusOK,
                "index",
                gin.H{
                    "title": "Error",
                    "repo" : api.Repository{},
                    "has": false,
                    "currtag": "",
                    "errorinfo": fmt.Sprintf("Get Repository Error,[%s]", err.Error()),
                },
            )
            return
        }
    }else {
        if err := h.DB.
        Find(&result).Error; err != nil {
            c.HTML(http.StatusOK,
                "index",
                gin.H{
                    "title":    "Error",
                    "repo" :    api.Repository{},
                    "has":      false,
                    "currtag":  "",
                    "errorinfo": fmt.Sprintf("Get Repository Error,[%s]", err.Error()),
                },
            )
            return
        }
    }
    pg.Data = h.groupRepos(result)
    c.HTML(http.StatusOK, "index",
        gin.H{
            "title":     "REPOSITORY",
            "has":len(result) > 0,
            "page":pg,
        })
    return
}

func (h * WebHandler) DeleteTag(c *gin.Context) {
    stag, suc := c.GetQuery(QUERY_TAG)
    if !suc {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title": "Error",
                "repo" : api.Repository{},
                "has": false,
                "currtag": "",
                "errorinfo": "error parameter tag name needed by it=?",
            },
        )
        return
    }
    repoName, suc := c.GetQuery(QUERY_REPOSITORY)
    if !suc {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title": "Error",
                "repo" : api.Repository{},
                "has": false,
                "currtag": "",
                "errorinfo": fmt.Sprintf("error parameter repository name needed by it"),
            },
        )
        return
    }
    repo, err := h.getRepository(repoName)
    if err != nil {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title":     "Error",
                "repo" :     api.Repository{},
                "has":       false,
                "currtag":   "",
                "errorinfo": fmt.Sprintf("error get Repository by it=?[%s]", err.Error()),
            },
        )
        return
    }

    tag := h.getSelectedTag(repo.Tags, stag)
    if tag == nil {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title":     "Error",
                "repo" :     api.Repository{},
                "has":       false,
                "currtag":   "",
                "errorinfo": fmt.Sprintf("No curresponed Tag found by it=?[%s]", tag),
            },
        )
        return
    }
    if err := h.deleteTag(repoName, tag.Digest); err != nil {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title":     "Error",
                "repo" :     api.Repository{},
                "has":       false,
                "currtag":   "",
                "errorinfo": fmt.Sprintf("error delete Tag[%s][%s][%s] by it?[%s]",repoName, tag.Name, tag.Digest,err.Error()),
            },
        )
        return
    }
    if err := h.DB.Delete(tag).Error; err != nil {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title":     "Error",
                "repo" :     api.Repository{},
                "has":       false,
                "currtag":   "",
                "errorinfo": fmt.Sprintf("error delete Tag[%s] by it?[%s]", tag.Name, err.Error()),
            },
        )
        return
    }
    h.GetTag(c)
}
func (h *WebHandler) GetTag(c *gin.Context) {
    repoName, suc := c.GetQuery(QUERY_REPOSITORY)
    if !suc {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title":     "Error",
                "repo" :     api.Repository{},
                "has":       false,
                "currtag":   "",
                "errorinfo": "error parameter repository name needed by it=?",
            },
        )
        return
    }
    repo, err := h.getRepository(repoName)
    if err != nil {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title":     "Error",
                "repo" :     api.Repository{},
                "has":       false,
                "currtag":   "",
                "errorinfo": fmt.Sprintf("error get repository name needed by it=?,[%s]", err.Error()),
            },
        )
        return
    }
    //p, _ := c.GetQuery(QUERY_TAG)
    //fmt.Printf("%+v::::::::::::::::%+v:::::::::::::: %+v\n",p,repoName,repo)
    //tag := h.getSelectedTag(repo.Tags, p)
    //if tag == nil {
    //    c.HTML(http.StatusOK,
    //        "contags",
    //        gin.H{
    //            "title":     "Error",
    //            "repo" :     api.Repository{},
    //            "has":       false,
    //            "currtag":   "",
    //            "errorinfo": "No Tags Found!",
    //        },
    //    )
    //    return
    //}
    //
    //fmt.Printf("TAGS: %+v \n", tag)
    c.HTML(http.StatusOK, "contags",
        gin.H{
            "title":     "REPOSITORY",
            "repo" : repo,
            "has": len(repo.Tags) > 0,
            "currtag":"woca",
        })
    return
}


func (h *WebHandler) PutTag(c *gin.Context) {
    stag, suc := c.GetQuery(QUERY_TAG)
    if !suc {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title":     "Error",
                "repo" :     api.Repository{},
                "has":       false,
                "currtag":   "",
                "errorinfo": "error parameter tag name needed by it=?",
            },
        )
        return
    }
    txt, suc := c.GetPostForm("txtbody-" + stag)
    if !suc {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title":     "Error",
                "repo" :     api.Repository{},
                "has":       false,
                "currtag":   "",
                "errorinfo": fmt.Sprintf("error parameter txtbody name needed by it=%s", stag),
            },
        )
        return
    }
    repoName, suc := c.GetQuery(QUERY_REPOSITORY)
    if !suc {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title":     "Error",
                "repo" :     api.Repository{},
                "has":       false,
                "currtag":   "",
                "errorinfo": fmt.Sprintf("error parameter repository name needed by it"),
            },
        )
        return
    }
    repo, err := h.getRepository(repoName)
    if err != nil {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title":     "Error",
                "repo" :     api.Repository{},
                "has":       false,
                "currtag":   "",
                "errorinfo": fmt.Sprintf("error get Repository by it=?[%s]", err.Error()),
            },
        )
        return
    }

    tag := h.getSelectedTag(repo.Tags, stag)
    if tag == nil {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title":     "Error",
                "repo" :     api.Repository{},
                "has":       false,
                "currtag":   "",
                "errorinfo": fmt.Sprintf("No curresponed Tag found by it=?[%s]", tag),
            },
        )
        return
    }
    tag.Description = txt
    if err := h.DB.Save(tag).Error; err != nil {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title":     "Error",
                "repo" :     api.Repository{},
                "has":       false,
                "currtag":   "",
                "errorinfo": fmt.Sprintf("error save Description by it=?[%s]", err.Error()),
            },
        )
        return
    }
    repo, err = h.getRepository(repoName)
    if err != nil {
        c.HTML(http.StatusOK,
            "contags",
            gin.H{
                "title":     "Error",
                "repo" :     api.Repository{},
                "has":       false,
                "currtag":   "",
                "errorinfo": fmt.Sprintf("error get Repository by it=?[%s]", err.Error()),
            },
        )
        return
    }
    c.HTML(http.StatusOK, "contags",
        gin.H{
            "title":     "REPOSITORY",
            "repo" : repo,
            "has": len(repo.Tags) > 0,
            "currtag":tag,
        })
    return
}

func (h *WebHandler) Help(c *gin.Context) {
    c.HTML(http.StatusOK, "help", gin.H{
        "title": "Sigma Help",
    })
}


func (h *WebHandler) deleteTag(repo, dgst string)error{
    ctx   := context.Background()
    repon,err := reference.ParseNamed(repo)
    if err != nil{
        return err
    }
    rp, err := client.NewRepository(ctx, repon, h.RegURL, nil)

    if err != nil {
        return err
    }
    ts,err := rp.Manifests(ctx)
    if err != nil{
        return err
    }
    d,e := digest.ParseDigest(dgst)
    if e != nil{
        return e
    }
    return ts.Delete(ctx,d)
}

func (h *WebHandler) getRepository(repoName string) (*api.Repository, error) {
    var tags []api.Tag
    repo := api.Repository{RepoName:repoName}
    if err := h.DB.Where("repo_name = ?", repoName).First(&repo).Error; err != nil {
        fmt.Errorf("Find repository [%s] Error,[%+v]", repoName, err)
        return nil, err
    }
    if err := h.DB.Model(&repo).Related(&tags).Error; err != nil {
        fmt.Errorf("Find Tags [%s] Error,[%+v]", repoName, err)
        return nil, err
    }
    repo.Tags = tags
    return &repo, nil
}
func (h *WebHandler) getSelectedTag(tags []api.Tag, tag string) *api.Tag {
    var ctag api.Tag
    if tag == "" {
        //return default first one
        if len(tags) > 0 {
            return &tags[0]
        }
        return nil
    }
    // Search for right tag
    var found bool
    for _, i := range (tags) {
        if tag == i.Name {
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

func (h *WebHandler) getPager(idx string) *Pager {
    if p, err := strconv.Atoi(idx); err == nil {
        if p > 1 {
            return &Pager{
                Size:PAGE_SIZE,
                Index:      p,
                Previous:   p -1,
                Next:       p +1,
            }
        }
    }
    return &Pager{
        Size:PAGE_SIZE,
        Index:      1,
        Previous:   1,
        Next:       2,
    }
}

func (h *WebHandler) groupRepos(rs []api.Repository) map[string][]api.Repository {
    se := make(map[string][]api.Repository)
    for _, it := range (rs) {
        key, exist := se[it.Group]
        if !exist {
            key = []api.Repository{it}
        }else {
            key = append(key, it)
        }
        se[it.Group] = key
    }
    fmt.Printf("%+v\n", se)
    //    var result [][]api.Repository
    //    for it,_ := range(se){
    //        fmt.Printf("%+v\n",it)
    //        result = append(result,it)
    //    }
    return se
}
