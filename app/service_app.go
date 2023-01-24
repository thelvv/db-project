package app

import (
	"forum/domain/entity"
	"forum/domain/repository"
)

type ServiceApp struct {
	s repository.ServiceRepository
}

func NewServiceApp(s repository.ServiceRepository) *ServiceApp {
	return &ServiceApp{s: s}
}

type ServiceAppInterface interface {
	ClearAllDate() error
	GetDBStatus() (*entity.Status, error)
}

func (s *ServiceApp) ClearAllDate() error {
	return s.s.ClearAllDate()
}

func (s *ServiceApp) GetDBStatus() (*entity.Status, error) {
	return s.s.GetDBStatus()
}
