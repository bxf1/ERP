package repository

import (
	"gorm.io/gorm"
	"github.com/bxf1/ERP/backend/pkg/rag/model"
)

type UserRepository struct {
	*BaseRepository
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *UserRepository) FindAll() ([]model.User, error) {
	var users []model.User
	err := r.DB().Find(&users).Error
	return users, err
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.DB().First(&user, id).Error
	return &user, err
}

func (r *UserRepository) Create(user *model.User) error {
	return r.DB().Create(user).Error
}

func (r *UserRepository) Update(user *model.User) error {
	return r.DB().Save(user).Error
}

func (r *UserRepository) Delete(id uint) error {
	return r.DB().Delete(&model.User{}, id).Error
}
