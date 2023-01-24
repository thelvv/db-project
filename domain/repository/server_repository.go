package repository

import "forum/domain/entity"

type ServiceRepository interface {
	ClearAllDate() error
	GetDBStatus() (*entity.Status, error)
}
