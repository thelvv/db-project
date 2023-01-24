package repository

import "forum/domain/entity"

type ThreadRepository interface {
	CreatePosts(thread *entity.Thread, posts []entity.Post) error
	CreateThread(thread *entity.Thread) error
	GetThreadPosts(slug string, limit int32, since string, order string) ([]entity.Post, error)
	GetThreadPostsTree(slug string, limit int32, since string, order string) ([]entity.Post, error)
	GetThreadPostsParentTree(slug string, limit int32, since string, order string) ([]entity.Post, error)
	CheckThreadBySlug(slug string) (int, error)
	GetThreadForumAndID(slugOrID string) (*entity.Thread, error)
	GetThreadsByForumSlug(slug string, limit int32, since string, desc bool) ([]entity.Thread, error)
	CheckThreadByID(ID int) error
	VoteForThread(vote *entity.Vote) (*entity.Thread, error)
	GetThreadBySlug(slug string) (*entity.Thread, error)
	GetThreadByID(ID int) (*entity.Thread, error)
	UpdateThread(thread *entity.Thread) error
}
