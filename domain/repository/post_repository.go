package repository

import "forum/domain/entity"

type PostRepository interface {
	GetPostDetails(postID int) (*entity.Post, error)
	ChangePostMessage(post *entity.Post) (*entity.Post, error)
}
