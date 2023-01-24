package infrastructure

import (
	"context"
	"fmt"
	"forum/domain/entity"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ForumRepo struct {
	db *pgxpool.Pool
}

func NewForumRepository(db *pgxpool.Pool) *ForumRepo {
	return &ForumRepo{db}
}

const CreateForumQuery = `INSERT INTO forums (slug, title, user_nickname) VALUES($1, $2, $3)`

func (f *ForumRepo) CreateForum(forumInput *entity.Forum) error {
	_, err := f.db.Exec(context.Background(), CreateForumQuery, forumInput.Slug, forumInput.Title, forumInput.User)
	return err
}

const GetForumDetailsQuery = `SELECT slug, title, user_nickname, thread_count, post_count FROM forums WHERE slug = $1`

func (f *ForumRepo) GetForumDetails(slug string) (*entity.Forum, error) {
	forum := &entity.Forum{}

	err := f.db.QueryRow(context.Background(), GetForumDetailsQuery, slug).Scan(
		&forum.Slug,
		&forum.Title,
		&forum.User,
		&forum.Threads,
		&forum.Posts)

	if err != nil {
		return nil, err
	}
	return forum, nil
}

func (f *ForumRepo) GetForumUsers(slug string, limit int32, since string, order string, compare string) ([]entity.User, error) {
	var query string
	if since != "" {
		if limit != 0 {
			query = fmt.Sprintf(`SELECT u.about, u.email, u.fullname, u.nickname FROM users AS u
				JOIN forum_user AS fu ON u.nickname = fu.nickname
				WHERE fu.forum_slug = '%s' AND fu.nickname %v '%s'
				ORDER BY u.nickname %v
				LIMIT %v`, slug, compare, since, order, limit)
		} else {

			query = fmt.Sprintf(`SELECT u.about, u.email, u.fullname, u.nickname FROM users AS u
				JOIN forum_user AS fu ON u.nickname = fu.nickname
				WHERE fu.forum_slug = '%s' AND fu.nickname %v '%s'
				ORDER BY u.nickname %v`, slug, compare, since, order)
		}
	} else {
		if limit != 0 {
			query = fmt.Sprintf(`SELECT u.about, u.email, u.fullname, u.nickname FROM users AS u
				JOIN forum_user AS fu ON u.nickname = fu.nickname
				WHERE fu.forum_slug = '%s'
				ORDER BY u.nickname %v
				LIMIT %v`, slug, order, limit)
		} else {
			query = fmt.Sprintf(`SELECT u.about, u.email, u.fullname, u.nickname FROM users AS u
				JOIN forum_user AS fu ON u.nickname = fu.nickname
				WHERE fu.forum_slug = '%s' 
				ORDER BY u.nickname %v`, slug, order)
		}
	}

	rows, err := f.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]entity.User, 0, limit)
	for rows.Next() {
		user := entity.User{}
		err = rows.Scan(&user.About, &user.Email, &user.Fullname, &user.Nickname)
		if err != nil {
			return nil, err // TODO: error handling
		}
		users = append(users, user)
	}
	return users, nil
}

const CheckForumQuery = `SELECT slug FROM forums WHERE slug = $1`

func (f *ForumRepo) CheckForum(slug string) (string, error) {
	err := f.db.QueryRow(context.Background(), CheckForumQuery, slug).Scan(&slug)

	if err != nil {
		return "", err
	}

	return slug, nil
}
