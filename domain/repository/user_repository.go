package repository

import "forum/domain/entity"

type UserRepository interface {
	CreateUser(user *entity.User) error
	CheckIfUserExists(nickname string) (string, error)
	GetUserByNickname(nickname string) (*entity.User, error)
	UpdateUser(newUser *entity.User) (*entity.User, error)
	GetUserNicknameWithEmail(email string) (string, error)
	GetUsersWithNicknameAndEmail(nickname, email string) ([]entity.User, error)
}
