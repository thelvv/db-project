package app

import (
	"fmt"
	"forum/domain/entity"
	"forum/domain/repository"
)

type ForumApp struct {
	f repository.ForumRepository
}

func NewForumApp(f repository.ForumRepository) *ForumApp {
	return &ForumApp{f: f}
}

type ForumAppInterface interface {
	CreateForum(forumInput *entity.Forum) error
	GetForumDetails(slug string) (*entity.Forum, error)
	GetForumUsers(slug string, limit int32, since string, desc bool) ([]entity.User, error)
	CheckForumCase(slug string) (string, error)
}

func (f *ForumApp) CreateForum(forumInput *entity.Forum) error {
	return f.f.CreateForum(forumInput)
}

func (f *ForumApp) GetForumDetails(slug string) (*entity.Forum, error) {
	return f.f.GetForumDetails(slug)
}

func (f *ForumApp) GetForumUsers(slug string, limit int32, since string, desc bool) ([]entity.User, error) {
	order := "ASC"
	var compare string
	if desc {
		order = "DESC"
		compare = "<"
	} else {
		compare = ">"
	}
	fmt.Println("=====>", slug, "LIMIT: ", limit, "SINCE: ", since, "DESC: ", desc, "<=====")
	return f.f.GetForumUsers(slug, limit, since, order, compare)
}

func (f *ForumApp) CheckForumCase(slug string) (string, error) {
	return f.f.CheckForum(slug)
}
