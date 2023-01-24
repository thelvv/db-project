package repository

import "forum/domain/entity"

type ForumRepository interface {
	CreateForum(forumInput *entity.Forum) error
	GetForumDetails(slug string) (*entity.Forum, error)
	GetForumUsers(slug string, limit int32, since string, order string, compare string) ([]entity.User, error)
	CheckForum(slug string) (string, error)
}
