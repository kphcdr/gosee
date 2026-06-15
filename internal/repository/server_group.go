package repository

import (
	"gorm.io/gorm"

	"gosee/internal/model"
)

type ServerGroupRepository struct {
	db *gorm.DB
}

func NewServerGroupRepository(db *gorm.DB) *ServerGroupRepository {
	return &ServerGroupRepository{db: db}
}

// List 查询全部分组（数量不多，不分页）
func (r *ServerGroupRepository) List(keyword string) ([]model.ServerGroup, error) {
	var groups []model.ServerGroup
	tx := r.db.Model(&model.ServerGroup{})
	if keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("name LIKE ?", like)
	}
	err := tx.Order("id DESC").Find(&groups).Error
	return groups, err
}

func (r *ServerGroupRepository) FindByID(id int64) (*model.ServerGroup, error) {
	var g model.ServerGroup
	if err := r.db.First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *ServerGroupRepository) Create(g *model.ServerGroup) error {
	return r.db.Create(g).Error
}

func (r *ServerGroupRepository) Update(g *model.ServerGroup) error {
	return r.db.Save(g).Error
}

func (r *ServerGroupRepository) Delete(id int64) error {
	return r.db.Delete(&model.ServerGroup{}, id).Error
}

// CountServers 统计分组下服务器数量
func (r *ServerGroupRepository) CountServers(groupID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Server{}).Where("group_id = ?", groupID).Count(&count).Error
	return count, err
}
