package api

import (
    "github.com/jinzhu/gorm"
    "time"
)

type Repository struct {
    gorm.Model
    RepoName string
    Group    string
    Tags     []Tag
    TPullCount uint
}

type Tag struct {
    gorm.Model
    RepositoryID uint `sql:"index"`
    Name         string
    Description  string `sql:"type:varchar(255)"`
    Digest       string
    PushTime     time.Time
    PushTimeEX   string `gorm:"-"`
    UpdatedAtEX  string `gorm:"-"`
    PullCount    uint
    Size         int64

    Layers       []Layer `gorm:"-"`
}

type Layer struct {

    CreatedBy   string
    Color       string
    Author      string
    Created     time.Time
}

