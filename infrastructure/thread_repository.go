package infrastructure

import (
	"context"
	"fmt"
	"forum/domain/entity"
	"github.com/go-openapi/strfmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"strconv"
	"time"
)

type ThreadRepo struct {
	db *pgxpool.Pool
}

func NewThreadRepository(db *pgxpool.Pool) *ThreadRepo {
	return &ThreadRepo{db: db}
}

const UpdatePostsCountQuery = `UPDATE forums SET post_count = post_count + $1 WHERE slug = $2;`
const GetThreadFromPostsQuery = `SELECT thread FROM posts WHERE id = $1`
const SelectSlugFromThread = `SELECT forum FROM threads WHERE id = $1`

func (t *ThreadRepo) CreatePosts(thread *entity.Thread, posts []entity.Post) error {
	var CreatePostsQuery = `INSERT INTO posts(author, created, forum, msg, parent, thread) VALUES `
	if posts[0].Parent != 0 {
		var parentThread int
		err := t.db.QueryRow(context.Background(), GetThreadFromPostsQuery, posts[0].Parent).Scan(&parentThread)

		if err != nil {
			return err
		}

		if parentThread != thread.ID {
			return entity.WrongParentError
		}
	}

	var postArray []interface{}
	created := strfmt.DateTime(time.Now())

	for i, post := range posts {
		posts[i].Forum = thread.Forum
		posts[i].Thread = thread.ID
		posts[i].Created = created

		CreatePostsQuery += fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d),",
			i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6,
		)

		postArray = append(postArray, post.Author, created, thread.Forum, post.Message, post.Parent, thread.ID)
	}

	CreatePostsQuery = CreatePostsQuery[:len(CreatePostsQuery)-1]
	CreatePostsQuery += ` RETURNING id`
	rows, err := t.db.Query(context.Background(), CreatePostsQuery, postArray...)
	if err != nil {
		return err
	}
	defer rows.Close()
	var idx int
	for rows.Next() {
		err = rows.Scan(&posts[idx].ID)
		if err != nil {
			return err
		}

		idx++
	}

	return nil
}

const CreateThreadQuery = `INSERT INTO threads (author, created, forum, msg, title, slug)
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

func (t *ThreadRepo) CreateThread(thread *entity.Thread) error {
	err := t.db.QueryRow(context.Background(), CreateThreadQuery,
		thread.Author, thread.Created, thread.Forum, thread.Message, thread.Title, thread.Slug,
	).Scan(&thread.ID)

	if err != nil {
		return err
	}

	return nil
}

func (t *ThreadRepo) GetThreadPosts(slug string, limit int32, since string, order string) ([]entity.Post, error) {
	var sinceQuery string
	if since != "" {
		if order == "DESC" {
			sinceQuery = fmt.Sprintf("AND id < %v", since)
		} else {
			sinceQuery = fmt.Sprintf("AND id > %v", since)
		}
	}

	threadID, err := strconv.Atoi(slug)
	if err != nil {
		threadID, err = t.CheckThreadBySlug(slug)
		if err != nil {
			return nil, err
		}
	}

	query := fmt.Sprintf(`SELECT author, created, forum, id, msg, parent, thread FROM posts
	WHERE thread = $1 %v
	ORDER BY id %v`, sinceQuery, order)

	if limit != 0 {
		query += fmt.Sprintf(" LIMIT %v", limit)
	}

	rows, err := t.db.Query(context.Background(), query, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]entity.Post, 0, limit)
	for rows.Next() {
		post := entity.Post{}
		err = rows.Scan(&post.Author, &post.Created, &post.Forum, &post.ID, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (t *ThreadRepo) GetThreadPostsTree(slug string, limit int32, since string, order string) ([]entity.Post, error) {
	var desc bool
	if order == "DESC" {
		desc = true
	} else {
		desc = false
	}

	threadID, err := strconv.Atoi(slug)
	if err != nil {
		threadID, err = t.CheckThreadBySlug(slug)
		if err != nil {
			return nil, err
		}
	}

	var query string

	if since == "" {
		if desc {
			query = fmt.Sprintf(`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE thread = %d ORDER BY path DESC, id  DESC LIMIT %d;`, threadID, limit)
		} else {
			query = fmt.Sprintf(`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE thread = %d ORDER BY path ASC, id  ASC LIMIT %d;`, threadID, limit)
		}
	} else {
		if desc {
			query = fmt.Sprintf(`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE thread = %d AND path < (SELECT path FROM posts WHERE id = %s)
				ORDER BY path DESC, id  DESC LIMIT %d;`, threadID, since, limit)
		} else {
			query = fmt.Sprintf(`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE thread = %d AND path > (SELECT path FROM posts WHERE id = %s)
				ORDER BY path ASC, id  ASC LIMIT %d;`, threadID, since, limit)
		}
	}

	rows, err := t.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]entity.Post, 0)
	for rows.Next() {
		post := entity.Post{}
		err = rows.Scan(&post.Author, &post.Created, &post.Forum, &post.ID, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (t *ThreadRepo) GetThreadPostsParentTree(slug string, limit int32, since string, order string) ([]entity.Post, error) {
	var desc bool
	if order == "DESC" {
		desc = true
	} else {
		desc = false
	}

	threadID, err := strconv.Atoi(slug)
	if err != nil {
		threadID, err = t.CheckThreadBySlug(slug)
		if err != nil {
			return nil, err
		}
	}

	var query string
	if since == "" {
		if desc {
			query = fmt.Sprintf(`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE path[1] IN (SELECT id FROM posts WHERE thread = %d AND parent = 0 ORDER BY id DESC LIMIT %d)
				ORDER BY path[1] DESC, path, id;`, threadID, limit)
		} else {
			query = fmt.Sprintf(`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE path[1] IN (SELECT id FROM posts WHERE thread = %d AND parent = 0 ORDER BY id LIMIT %d)
				ORDER BY path, id;`, threadID, limit)
		}
	} else {
		if desc {
			query = fmt.Sprintf(`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE path[1] IN (SELECT id FROM posts WHERE thread = %d AND parent = 0 AND path[1] <
				(SELECT path[1] FROM posts WHERE id = %s) ORDER BY id DESC LIMIT %d) ORDER BY path[1] DESC, path, id;`,
				threadID, since, limit)
		} else {
			query = fmt.Sprintf(`SELECT author, created, forum, id, msg, parent, thread FROM posts
				WHERE path[1] IN (SELECT id FROM posts WHERE thread = %d AND parent = 0 AND path[1] >
				(SELECT path[1] FROM posts WHERE id = %s) ORDER BY id ASC LIMIT %d) ORDER BY path, id;`,
				threadID, since, limit)
		}
	}

	rows, err := t.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	posts := make([]entity.Post, 0)
	for rows.Next() {
		post := entity.Post{}
		err = rows.Scan(&post.Author, &post.Created, &post.Forum, &post.ID, &post.Message, &post.Parent, &post.Thread)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

const CheckThreadBySlugQuery = `SELECT id FROM threads WHERE slug = $1`

func (t *ThreadRepo) CheckThreadBySlug(slug string) (int, error) {
	var id int
	err := t.db.QueryRow(context.Background(), CheckThreadBySlugQuery, slug).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

const CheckThreadByIDQuery = `SELECT id FROM threads WHERE id = $1`

func (t *ThreadRepo) CheckThreadByID(ID int) error {
	err := t.db.QueryRow(context.Background(), CheckThreadByIDQuery, ID).Scan(&ID)

	if err != nil {
		return err
	}

	return nil
}

const GetThreadForumAndIDBySlugQuery = `SELECT forum, id FROM threads WHERE slug = $1`
const GetThreadForumAndIDByIDQuery = `SELECT forum FROM threads WHERE id = $1`

func (t *ThreadRepo) GetThreadForumAndID(slugOrID string) (*entity.Thread, error) {
	threadID, err := strconv.Atoi(slugOrID)
	thread := &entity.Thread{ID: threadID}
	if err != nil {
		err = t.db.QueryRow(context.Background(), GetThreadForumAndIDBySlugQuery, slugOrID).Scan(&thread.Forum, &thread.ID)
	} else {
		err = t.db.QueryRow(context.Background(), GetThreadForumAndIDByIDQuery, thread.ID).Scan(&thread.Forum)
	}

	if err != nil {
		return nil, err
	}

	return thread, nil
}

func (t *ThreadRepo) GetThreadsByForumSlug(slug string, limit int32, since string, desc bool) ([]entity.Thread, error) {
	var GetThreadsByForumSlugQuery = `SELECT author, created, forum, id, msg, slug, title, votes FROM threads WHERE forum = $1`
	order := "ASC"
	var compare string
	if desc == false {
		compare = ">"
	} else {
		order = "DESC"
		compare = "<"
	}

	if since != "" {
		GetThreadsByForumSlugQuery += fmt.Sprintf(" AND created %v= $2", compare)
	}

	GetThreadsByForumSlugQuery += fmt.Sprintf(" ORDER BY created %v  LIMIT %v", order, limit)
	var rows pgx.Rows
	var err error
	if since != "" {
		rows, err = t.db.Query(context.Background(), GetThreadsByForumSlugQuery, slug, since)
	} else {
		rows, err = t.db.Query(context.Background(), GetThreadsByForumSlugQuery, slug)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	threads := make([]entity.Thread, 0, limit)
	for rows.Next() {
		thread := entity.Thread{}
		err = rows.Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
		if err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}

	return threads, nil
}

const GetVoteQuery = `SELECT vote FROM thread_vote WHERE nickname = $1 AND thread_id = $2`
const InsertVoteQuery = `INSERT INTO thread_vote (nickname, thread_id, vote) VALUES($1, $2, $3)`
const UpdateVoteQuery = `UPDATE thread_vote SET vote = $1 WHERE nickname = $2 AND thread_id = $3`

func (t *ThreadRepo) VoteForThread(vote *entity.Vote) (*entity.Thread, error) {
	thread := &entity.Thread{}
	var err error

	if vote.ID != 0 {
		thread, err = t.GetThreadByID(vote.ID)
	} else {
		thread, err = t.GetThreadBySlug(vote.Slug)
	}

	if err != nil {
		return nil, err
	}
	var voteValue int
	err = t.db.QueryRow(context.Background(), GetVoteQuery, vote.Nickname, thread.ID).Scan(&voteValue)

	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	if err == pgx.ErrNoRows {
		_, err = t.db.Exec(context.Background(), InsertVoteQuery, vote.Nickname, thread.ID, vote.Voice)

		if err != nil {
			return nil, err
		}

		thread.Votes += vote.Voice
		return thread, nil
	}

	if voteValue == vote.Voice {
		return thread, nil
	}

	thread.Votes = thread.Votes - voteValue + vote.Voice

	_, err = t.db.Exec(context.Background(), UpdateVoteQuery, vote.Voice, vote.Nickname, thread.ID)
	if err != nil {
		return nil, err
	}
	return thread, nil
}

const GetThreadBySlugQuery = `SELECT author, created, forum, id, msg, slug, title, votes FROM threads WHERE slug = $1`

func (t *ThreadRepo) GetThreadBySlug(slug string) (*entity.Thread, error) {
	thread := &entity.Thread{}
	err := t.db.QueryRow(context.Background(), GetThreadBySlugQuery, slug).Scan(
		&thread.Author,
		&thread.Created,
		&thread.Forum,
		&thread.ID,
		&thread.Message,
		&thread.Slug,
		&thread.Title,
		&thread.Votes)

	if err != nil {
		return nil, err
	}
	return thread, nil
}

const GetThreadByIDQuery = `SELECT author, created, forum, id, msg, slug, title, votes FROM threads WHERE id = $1`

func (t *ThreadRepo) GetThreadByID(ID int) (*entity.Thread, error) {
	thread := &entity.Thread{}
	err := t.db.QueryRow(context.Background(), GetThreadByIDQuery, ID).Scan(
		&thread.Author,
		&thread.Created,
		&thread.Forum,
		&thread.ID,
		&thread.Message,
		&thread.Slug,
		&thread.Title,
		&thread.Votes)

	if err != nil {
		return nil, err
	}
	return thread, nil
}

const UpdateThreadQuery = `UPDATE threads SET title = $1, msg = $2
		WHERE slug = $3 OR id = $4
		RETURNING author, created, forum, id, msg, slug, title`

func (t *ThreadRepo) UpdateThread(thread *entity.Thread) error {
	if thread.Title == "" || thread.Message == "" {
		oldThread := &entity.Thread{}
		var err error

		if thread.ID != 0 {
			oldThread, err = t.GetThreadByID(thread.ID)
		} else {
			oldThread, err = t.GetThreadBySlug(*thread.Slug)
		}
		if err != nil {
			return err
		}

		if thread.Title == "" {
			thread.Title = oldThread.Title
		}

		if thread.Message == "" {
			thread.Message = oldThread.Message
		}
	}

	err := t.db.QueryRow(context.Background(), UpdateThreadQuery,
		thread.Title, thread.Message, thread.Slug, thread.ID,
	).Scan(&thread.Author, &thread.Created, &thread.Forum, &thread.ID, &thread.Message, &thread.Slug, &thread.Title)

	if err != nil {
		return err
	}

	return nil
}
