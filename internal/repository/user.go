package repository

import (
	"time"

	"gorm.io/gorm"

	"gosee/internal/model"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	var u model.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) UpdatePassword(id int64, hashed string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("password", hashed).Error
}

func (r *UserRepository) UpdateLastLogin(id int64) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("last_login_at", time.Now()).Error
}
