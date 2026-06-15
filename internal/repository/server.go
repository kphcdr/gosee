package repository

import (
	"gorm.io/gorm"

	"gosee/internal/model"
)

// ServerListQuery 服务器列表查询条件
type ServerListQuery struct {
	Page     int
	PageSize int
	GroupID  *int64
	Enabled  *int8
	Keyword  string
}

type ServerRepository struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

// List 分页查询服务器列表
func (r *ServerRepository) List(q ServerListQuery) ([]model.Server, int64, error) {
	var (
		servers []model.Server
		total   int64
	)
	tx := r.db.Model(&model.Server{})
	if q.GroupID != nil {
		tx = tx.Where("group_id = ?", *q.GroupID)
	}
	if q.Enabled != nil {
		tx = tx.Where("enabled = ?", *q.Enabled)
	}
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		tx = tx.Where("name LIKE ? OR host LIKE ?", like, like)
	}
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	off := (q.Page - 1) * q.PageSize
	if err := tx.Order("id DESC").Offset(off).Limit(q.PageSize).Find(&servers).Error; err != nil {
		return nil, 0, err
	}
	return servers, total, nil
}

// FindByID 按 ID 查询
func (r *ServerRepository) FindByID(id int64) (*model.Server, error) {
	var s model.Server
	if err := r.db.First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// FindEnabledByID 仅查启用的服务器
func (r *ServerRepository) FindEnabledByID(id int64) (*model.Server, error) {
	var s model.Server
	if err := r.db.Where("id = ? AND enabled = 1", id).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// ListEnabled 查询全部启用的服务器（供定时全量采集使用）
func (r *ServerRepository) ListEnabled() ([]model.Server, error) {
	var servers []model.Server
	err := r.db.Where("enabled = 1").Order("id ASC").Find(&servers).Error
	return servers, err
}

// Create 新增服务器
func (r *ServerRepository) Create(s *model.Server) error {
	return r.db.Create(s).Error
}

// Update 更新服务器（全字段）
func (r *ServerRepository) Update(s *model.Server) error {
	return r.db.Save(s).Error
}

// Delete 删除服务器
func (r *ServerRepository) Delete(id int64) error {
	return r.db.Delete(&model.Server{}, id).Error
}

// UpdateStatus 更新状态与最近采集信息
func (r *ServerRepository) UpdateStatus(id int64, status string, lastError string, checked bool) error {
	updates := map[string]interface{}{"status": status}
	if lastError != "" {
		updates["last_error"] = lastError
	} else {
		updates["last_error"] = nil
	}
	if checked {
		updates["last_checked_at"] = gorm.Expr("CURRENT_TIMESTAMP")
	}
	return r.db.Model(&model.Server{}).Where("id = ?", id).Updates(updates).Error
}
