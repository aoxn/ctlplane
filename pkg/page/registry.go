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
    "fmt"
    "errors"
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

type Manifest struct {
    RepoName string
    TagName string
    Digest  string
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


func ListAll() ([]api.Repository,error){
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
    var result []api.Repository
    for _,it := range(repo.Repos){
        tags,_:=ListTag(it)
        var nTag []api.Tag
        for _,t:= range(tags.Tags){
            if m,err := GetManifest(it,t);err == nil{
                nTag = append(nTag,api.Tag{Name:t,Digest:m.Digest})
                continue
            }
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

func GetManifest(repoName,tag string) (*Manifest,error){
    url := REPO_SERVER + "v2/"+repoName+"/manifests/"+tag
    res,err := http.Get(url)
    if err != nil{
        return nil,err
    }
    defer res.Body.Close()
    digest := res.Header.Get("Docker-Content-Digest")
    //fmt.Println("Digest: %s",digest)
    if digest == ""{
        fmt.Errorf("Warning: content digest nil !")
    }
    return &Manifest{RepoName:repoName,TagName:tag,Digest:digest},err
}

func DeleteTag(repo,digest string)(error){
    url := REPO_SERVER + "v2/"+repo+"/manifests/"+digest
    r := http.Client{}
    req,err := http.NewRequest("DELETE",url,nil)
    if err != nil{
        return err
    }
    res,err := r.Do(req)
    if err != nil{
        return err
    }
    defer res.Body.Close()
    if res.StatusCode != http.StatusAccepted{
        msg,err := ioutil.ReadAll(res.Body)
        if err != nil{
            msg = []byte("error read msg")
        }
        return errors.New(string(msg))
    }
    return err
}

