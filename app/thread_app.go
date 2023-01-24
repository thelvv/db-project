package app

import (
	"forum/domain/entity"
	"forum/domain/repository"
	"strconv"
)

type ThreadApp struct {
	t        repository.ThreadRepository
	forumApp ForumAppInterface
}

func NewThreadApp(f repository.ThreadRepository, forumApp ForumAppInterface) *ThreadApp {
	return &ThreadApp{t: f, forumApp: forumApp}
}

type ThreadAppInterface interface {
	CreatePosts(thread *entity.Thread, posts []entity.Post) error
	CreateThread(thread *entity.Thread) error
	GetThreadPosts(slug string, limit int32, since string, sort string, desc bool) ([]entity.Post, error)
	CheckThread(slugOrID string) error
	VoteForThread(vote *entity.Vote) (*entity.Thread, error)
	GetThread(slugOrID string) (*entity.Thread, error)
	GetThreadForumAndID(slugOrID string) (*entity.Thread, error)
	GetThreadsByForumSlug(slug string, limit int32, since string, desc bool) ([]entity.Thread, error)
	UpdateThread(slugOrID string, newThreadData *entity.Thread) error
}

func (t *ThreadApp) CreatePosts(thread *entity.Thread, posts []entity.Post) error {
	return t.t.CreatePosts(thread, posts)
}

func (t *ThreadApp) CreateThread(thread *entity.Thread) error {
	var err error
	thread.Forum, err = t.forumApp.CheckForumCase(thread.Forum)
	if err != nil {
		return entity.ForumNotExistError
	}
	return t.t.CreateThread(thread)
}

func (t *ThreadApp) GetThreadPosts(slug string, limit int32, since string, sort string, desc bool) ([]entity.Post, error) {
	order := "ASC"
	switch desc {
	case true:
		order = "DESC"
	}

	switch sort {
	case "flat":
		return t.t.GetThreadPosts(slug, limit, since, order)
	case "tree":
		return t.t.GetThreadPostsTree(slug, limit, since, order)
	case "parent_tree":
		return t.t.GetThreadPostsParentTree(slug, limit, since, order)
	default:
		return t.t.GetThreadPosts(slug, limit, since, order)
	}
}

func (t *ThreadApp) CheckThread(slugOrID string) error {
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		_, err = t.t.CheckThreadBySlug(slugOrID)
		return err
	}

	return t.t.CheckThreadByID(id)
}

func (t *ThreadApp) VoteForThread(vote *entity.Vote) (*entity.Thread, error) {
	return t.t.VoteForThread(vote)
}

func (t *ThreadApp) GetThread(slugOrID string) (*entity.Thread, error) {
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		return t.t.GetThreadBySlug(slugOrID)
	}
	return t.t.GetThreadByID(id)
}

func (t *ThreadApp) GetThreadForumAndID(slugOrID string) (*entity.Thread, error) {
	return t.t.GetThreadForumAndID(slugOrID)
}

func (t *ThreadApp) GetThreadsByForumSlug(slug string, limit int32, since string, desc bool) ([]entity.Thread, error) {
	return t.t.GetThreadsByForumSlug(slug, limit, since, desc)
}

func (t *ThreadApp) UpdateThread(slugOrID string, newThreadData *entity.Thread) error {
	newThreadData.Slug = &slugOrID
	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		id = 0
	}

	newThreadData.ID = id
	return t.t.UpdateThread(newThreadData)
}
