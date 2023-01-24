package infrastructure

import (
	"context"
	"forum/domain/entity"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PostRepo struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) *PostRepo {
	return &PostRepo{db: db}
}

const GetPostDetailsQuery = `SELECT author, created, forum, id, msg, thread, isEdited, parent FROM posts WHERE id = $1`

func (p *PostRepo) GetPostDetails(postID int) (*entity.Post, error) {
	post := &entity.Post{}
	err := p.db.QueryRow(context.Background(), GetPostDetailsQuery, postID).Scan(
		&post.Author,
		&post.Created,
		&post.Forum,
		&post.ID,
		&post.Message,
		&post.Thread,
		&post.IsEdited,
		&post.Parent)

	if err != nil {
		return nil, err
	}

	return post, nil
}

const ChangePostMessageQuery = `UPDATE posts SET msg = $1, isEdited = true 
	          WHERE id = $2
	          RETURNING author, created, forum, id, msg, thread, isEdited, parent`

func (p *PostRepo) ChangePostMessage(post *entity.Post) (*entity.Post, error) {
	err := p.db.QueryRow(context.Background(), ChangePostMessageQuery, post.Message, post.ID).Scan(
		&post.Author,
		&post.Created,
		&post.Forum,
		&post.ID,
		&post.Message,
		&post.Thread,
		&post.IsEdited,
		&post.Parent)

	if err != nil {
		return nil, err
	}
	return post, nil
}
