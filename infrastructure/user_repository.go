package infrastructure

import (
	"context"
	"forum/domain/entity"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db}
}

// emptyIfNil replaces nil input with pointer to empty string, noop otherwise
func emptyIfNil(input *string) *string {
	if input == nil {
		return new(string)
	}
	return input
}

const CreateUserQuery = `INSERT INTO users (nickname, fullname, email, about) VALUES ($1, $2, $3, $4)`

func (us *UserRepo) CreateUser(user *entity.User) error {
	_, err := us.db.Exec(context.Background(),
		CreateUserQuery,
		user.Nickname, user.Fullname, user.Email, user.About,
	)

	if err != nil {
		return err
	}

	return nil
}

const CheckUserExistQuery = `SELECT nickname FROM users WHERE nickname = $1`

func (us *UserRepo) CheckIfUserExists(nickname string) (string, error) {
	err := us.db.QueryRow(context.Background(), CheckUserExistQuery, nickname).Scan(&nickname)
	if err != nil {
		return "", err
	}
	return nickname, nil
}

const GetUserByNickname = `SELECT id, nickname, fullname, email, about FROM users WHERE nickname = $1`

func (us *UserRepo) GetUserByNickname(nickname string) (*entity.User, error) {
	user := &entity.User{}
	err := us.db.QueryRow(context.Background(), GetUserByNickname, nickname).Scan(
		&user.ID,
		&user.Nickname,
		&user.Fullname,
		&user.Email,
		&user.About)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.UserDoesntExistsError
		}
		return nil, err
	}

	return user, nil
}

const UpdateUserQuery = `UPDATE users SET fullname = $1, email = $2, about = $3 WHERE id = $4`

func (us *UserRepo) UpdateUser(newUser *entity.User) (*entity.User, error) {
	_, err := us.db.Exec(context.Background(), UpdateUserQuery, newUser.Fullname, newUser.Email, newUser.About, newUser.ID)
	if err != nil {
		return nil, entity.DataError
	}

	return newUser, nil
}

const GetUserNicknameWithEmailQuery = `SELECT nickname FROM users WHERE email = $1`

func (us *UserRepo) GetUserNicknameWithEmail(email string) (string, error) {
	var nickname string
	err := us.db.QueryRow(context.Background(), GetUserNicknameWithEmailQuery, email).Scan(&nickname)

	if err != nil {
		return "", err
	}

	return nickname, nil
}

const GetUserWithNicknameAndEmailQuery = `SELECT nickname, fullname, email, about FROM users
		WHERE nickname = $1 OR email = $2`

func (us *UserRepo) GetUsersWithNicknameAndEmail(nickname, email string) ([]entity.User, error) {
	rows, err := us.db.Query(context.Background(), GetUserWithNicknameAndEmailQuery, nickname, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]entity.User, 0)
	for rows.Next() {
		user := entity.User{}
		err = rows.Scan(&user.Nickname, &user.Fullname, &user.Email, &user.About)
		if err != nil {
			return nil, err // TODO: error handling
		}
		users = append(users, user)
	}

	return users, nil
}
