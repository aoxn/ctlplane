package api
import "github.com/jinzhu/gorm"


type Repos []Repository

type Repository struct {
	gorm.Model
	RepoName string
	Tags Tags
}

type Tags []Tag

type Tag struct{
	gorm.Model
	RepositoryID uint `sql:"index"`
	Name string
	Description string `sql:"type:varchar(255);unique_index"`
}

