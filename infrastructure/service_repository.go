package infrastructure

import (
	"context"
	"forum/domain/entity"
	"github.com/jackc/pgx/v4/pgxpool"
)

type ServiceRepo struct {
	db *pgxpool.Pool
}

func NewServiceRepository(db *pgxpool.Pool) *ServiceRepo {
	return &ServiceRepo{db: db}
}

const ClearDBQuery = `TRUNCATE TABLE Forum_user RESTART IDENTITY CASCADE;
			  TRUNCATE TABLE Thread_vote RESTART IDENTITY CASCADE;
			  TRUNCATE TABLE Posts RESTART IDENTITY CASCADE;
			  TRUNCATE TABLE Threads RESTART IDENTITY CASCADE;
			  TRUNCATE TABLE Forums RESTART IDENTITY CASCADE;
			  TRUNCATE TABLE Users RESTART IDENTITY CASCADE;`

func (s *ServiceRepo) ClearAllDate() error {
	_, err := s.db.Exec(context.Background(), ClearDBQuery)
	if err != nil {
		return err
	}
	return nil
}

const GetUserStatusQuery = `SELECT COUNT(*) AS user_count FROM Users;`
const GetThreadStatusQuery = `SELECT COUNT(*) AS thread_count FROM Threads;`
const GetForumStatusQuery = `SELECT COUNT(*) AS forum_count FROM Forums;`
const GetPostStatusQuery = `SELECT COUNT(*) AS post_count FROM Posts;`

func (s *ServiceRepo) GetDBStatus() (*entity.Status, error) {
	status := &entity.Status{}

	err := s.db.QueryRow(context.Background(), GetUserStatusQuery).Scan(&status.User)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRow(context.Background(), GetForumStatusQuery).Scan(&status.Forum)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRow(context.Background(), GetThreadStatusQuery).Scan(&status.Thread)
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRow(context.Background(), GetPostStatusQuery).Scan(&status.Post)
	if err != nil {
		return nil, err
	}

	return status, nil
}
