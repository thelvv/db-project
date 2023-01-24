package app

import (
	"forum/domain/entity"
	"forum/domain/repository"
)

type PostApp struct {
	p repository.PostRepository
}

func NewPostApp(p repository.PostRepository) *PostApp {
	return &PostApp{p: p}
}

type PostAppInterface interface {
	GetPostDetails(postID int) (*entity.Post, error)
	ChangePostMessage(post *entity.Post) (*entity.Post, error)
}

func (p *PostApp) GetPostDetails(postID int) (*entity.Post, error) {
	return p.p.GetPostDetails(postID)
}

func (p *PostApp) ChangePostMessage(post *entity.Post) (*entity.Post, error) {
	previousPost, err := p.GetPostDetails(post.ID)
	if err != nil {
		return nil, err
	}

	if post.Message == previousPost.Message {
		return previousPost, nil
	}
	return p.p.ChangePostMessage(post)
}
