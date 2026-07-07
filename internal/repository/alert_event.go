package repository

import (
	"time"

	"gorm.io/gorm"

	"gosee/internal/model"
)

type AlertEventRepository struct {
	db *gorm.DB
}

func NewAlertEventRepository(db *gorm.DB) *AlertEventRepository {
	return &AlertEventRepository{db: db}
}

// AlertEventView 事件 + 服务器名/规则名（JOIN 填充，供前端展示）
type AlertEventView struct {
	model.AlertEvent
	ServerName string `json:"server_name"`
	RuleName   string `json:"rule_name"`
}

// List 最近事件列表（按最近触发时间倒序），JOIN 填充 server/rule 名
func (r *AlertEventRepository) List(limit int) ([]AlertEventView, error) {
	return r.ListByGroup(limit, nil)
}

// ListByGroup 查询指定分组的最近事件；groupID 为空时查询全部。
func (r *AlertEventRepository) ListByGroup(limit int, groupID *int64) ([]AlertEventView, error) {
	return r.listByGroupSince(limit, groupID, nil)
}

// ListByGroupSince 查询指定分组在给定时间之后的最近事件。
func (r *AlertEventRepository) ListByGroupSince(limit int, groupID *int64, since time.Time) ([]AlertEventView, error) {
	return r.listByGroupSince(limit, groupID, &since)
}

func (r *AlertEventRepository) listByGroupSince(limit int, groupID *int64, since *time.Time) ([]AlertEventView, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	var rows []AlertEventView
	tx := r.db.Table("alert_events").
		Select("alert_events.*, servers.name AS server_name, alert_rules.name AS rule_name").
		Joins("LEFT JOIN servers ON servers.id = alert_events.server_id").
		Joins("LEFT JOIN alert_rules ON alert_rules.id = alert_events.alert_rule_id")
	if groupID != nil {
		tx = tx.Where("servers.group_id = ?", *groupID)
	}
	if since != nil {
		tx = tx.Where("alert_events.last_triggered_at >= ?", *since)
	}
	err := tx.
		Order("alert_events.last_triggered_at DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

// FindFiring 查某规则+服务器当前 firing 状态的事件（无则 ErrRecordNotFound）
func (r *AlertEventRepository) FindFiring(ruleID, serverID int64) (*model.AlertEvent, error) {
	var ev model.AlertEvent
	err := r.db.Where("alert_rule_id = ? AND server_id = ? AND status = ?",
		ruleID, serverID, model.EventStatusFiring).First(&ev).Error
	return &ev, err
}

func (r *AlertEventRepository) Create(ev *model.AlertEvent) error {
	return r.db.Create(ev).Error
}

// TouchFiring 更新已存在 firing 事件的最近触发时间与当前值
func (r *AlertEventRepository) TouchFiring(id int64, currentValue *float64) error {
	return r.db.Model(&model.AlertEvent{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_triggered_at": gorm.Expr("CURRENT_TIMESTAMP"),
			"current_value":     currentValue,
			"status":            model.EventStatusFiring,
		}).Error
}

// MarkRecovered 标记恢复
func (r *AlertEventRepository) MarkRecovered(id int64) error {
	return r.db.Model(&model.AlertEvent{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       model.EventStatusRecovered,
			"recovered_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}

// UpdateStatus 更新事件状态（ack / close）
func (r *AlertEventRepository) UpdateStatus(id int64, status string) error {
	return r.db.Model(&model.AlertEvent{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateNotified 记录已通知（防重复，阶段六 Notifier 调用）
func (r *AlertEventRepository) UpdateNotified(id int64) error {
	return r.db.Model(&model.AlertEvent{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_notified_at": gorm.Expr("CURRENT_TIMESTAMP"),
			"notify_count":     gorm.Expr("notify_count + 1"),
		}).Error
}

// DeleteEndedBefore 删除 status<>firing 且 updated_at 早于 t 的告警事件（绝不删正在触发的事件）。
// 分批处理，返回累计删除条数。调用方应在此之前先清理 alert_notifications，避免孤儿通知记录。
func (r *AlertEventRepository) DeleteEndedBefore(t time.Time, batch int) (int64, error) {
	if batch <= 0 {
		batch = 1000
	}
	var total int64
	for {
		var ids []int64
		if err := r.db.Model(&model.AlertEvent{}).
			Where("status <> ? AND updated_at < ?", model.EventStatusFiring, t).
			Limit(batch).Pluck("id", &ids).Error; err != nil {
			return total, err
		}
		if len(ids) == 0 {
			break
		}
		res := r.db.Where("id IN ?", ids).Delete(&model.AlertEvent{})
		if res.Error != nil {
			return total, res.Error
		}
		total += res.RowsAffected
		if len(ids) < batch {
			break
		}
	}
	return total, nil
}
