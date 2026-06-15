package repository

import (
	"gorm.io/gorm"

	"gosee/internal/model"
)

type AlertRuleRepository struct {
	db *gorm.DB
}

func NewAlertRuleRepository(db *gorm.DB) *AlertRuleRepository {
	return &AlertRuleRepository{db: db}
}

func (r *AlertRuleRepository) List() ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := r.db.Order("id ASC").Find(&rules).Error
	return rules, err
}

func (r *AlertRuleRepository) FindByID(id int64) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := r.db.First(&rule, id).Error
	return &rule, err
}

func (r *AlertRuleRepository) Create(rule *model.AlertRule) error {
	return r.db.Create(rule).Error
}

func (r *AlertRuleRepository) Update(rule *model.AlertRule) error {
	return r.db.Save(rule).Error
}

func (r *AlertRuleRepository) Delete(id int64) error {
	return r.db.Delete(&model.AlertRule{}, id).Error
}

func (r *AlertRuleRepository) SetEnabled(id int64, enabled int8) error {
	return r.db.Model(&model.AlertRule{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// ApplicableRules 查询某台服务器适用的启用规则：
// global 规则 + scope=server 指向本机 + scope=group 指向所属分组（groupID 为 nil 时跳过分组规则）
func (r *AlertRuleRepository) ApplicableRules(serverID int64, groupID *int64) ([]model.AlertRule, error) {
	var gid int64
	if groupID != nil {
		gid = *groupID
	}
	var rules []model.AlertRule
	err := r.db.Where(
		"enabled = 1 AND (scope_type = ? OR (scope_type = ? AND scope_id = ?) OR (scope_type = ? AND scope_id = ?))",
		model.ScopeTypeGlobal, model.ScopeTypeServer, serverID, model.ScopeTypeGroup, gid,
	).Find(&rules).Error
	return rules, err
}
