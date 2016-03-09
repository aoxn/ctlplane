package page
import (
//    "net/http"
//    "io/ioutil"
//    "fmt"
//    "encoding/json"
//    "github.com/spacexnice/ctlplane/pkg/api"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "github.com/spacexnice/ctlplane/pkg/api"
)
const (
    REPO_SERVER = "http://61.160.36.122:8080/"
)

type Repos struct {
    Repos []string `json:"repositories,omitempty"`
}
type Tag struct {
    Name string `json:"name,omitempty"`
    Tags []string `json:"tags,omitempty"`
}

func List() (*Repos,error){
    repo := &Repos{}
    url := REPO_SERVER + "v2/_catalog"
    res,err := http.Get(url)
    if err != nil{
        return nil,err
    }
    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    if err != nil{
        return nil,err
    }
    err = json.Unmarshal(body,repo)
    if err != nil{
        return nil,err
    }
    return repo,nil
}


func ListAll() (api.Repos,error){
    repo := &Repos{}
    url := REPO_SERVER + "v2/_catalog"
    res,err := http.Get(url)
    if err != nil{
        return nil,err
    }
    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    if err != nil{
        return nil,err
    }
    err = json.Unmarshal(body,repo)
    if err != nil{
        return nil,err
    }
    var result api.Repos
    for _,it := range(repo.Repos){
        tags,_:=ListTag(it)
        nTag := api.Tags{}
        for _,t:= range(tags.Tags){
            nTag = append(nTag,api.Tag{Name:t})
        }
        result = append(result,api.Repository{RepoName:it,Tags:nTag})
    }
    return result,nil
}

func ListTag(repo string) (*Tag,error){
    tags := &Tag{}
    url := REPO_SERVER + "v2/"+repo+"/tags/list"
    res,err := http.Get(url)
    if err != nil{
        return nil,err
    }
    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    if err != nil{
        return nil,err
    }
    err = json.Unmarshal(body,tags)
    if err != nil{
        return nil,err
    }
    return tags,nil
}

type PageLister interface {
    List(index int,obj interface{}) interface{}

}

