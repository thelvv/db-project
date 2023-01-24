package app

import (
	"forum/domain/entity"
	"forum/domain/repository"
)

type UserApp struct {
	us repository.UserRepository
}

func NewUserApp(us repository.UserRepository) *UserApp {
	return &UserApp{us: us}
}

type UserAppInterface interface {
	CreateUser(user *entity.User) error
	CheckIfUserExists(nickname string) (string, error)
	GetUserByNickname(nickname string) (*entity.User, error)
	UpdateUser(newUser *entity.User) (*entity.User, error)
	GetUserNicknameWithEmail(email string) (string, error)
	GetUsersWithNicknameAndEmail(nickname, email string) ([]entity.User, error)
}

func (us *UserApp) CreateUser(user *entity.User) error {
	return us.us.CreateUser(user)
}

func (us *UserApp) CheckIfUserExists(nickname string) (string, error) {
	return us.us.CheckIfUserExists(nickname)
}

func (us *UserApp) GetUserByNickname(nickname string) (*entity.User, error) {
	return us.us.GetUserByNickname(nickname)
}

func (us *UserApp) UpdateUser(newUser *entity.User) (*entity.User, error) {
	userFromDB, err := us.GetUserByNickname(newUser.Nickname)
	if err != nil {
		return nil, err
	}
	newUser.ID = userFromDB.ID
	if newUser.Fullname == "" {
		newUser.Fullname = userFromDB.Fullname
	}

	if newUser.Email == "" {
		newUser.Email = userFromDB.Email
	}

	if newUser.About == "" {
		newUser.About = userFromDB.About
	}

	return us.us.UpdateUser(newUser)
}

func (us *UserApp) GetUserNicknameWithEmail(email string) (string, error) {
	return us.us.GetUserNicknameWithEmail(email)
}

func (us *UserApp) GetUsersWithNicknameAndEmail(nickname, email string) ([]entity.User, error) {
	return us.us.GetUsersWithNicknameAndEmail(nickname, email)
}
