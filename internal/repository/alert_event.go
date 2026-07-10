package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

// FindFiring 查某规则+服务器当前活动事件（无则 ErrRecordNotFound）。
func (r *AlertEventRepository) FindFiring(ruleID, serverID int64) (*model.AlertEvent, error) {
	var ev model.AlertEvent
	err := r.db.Where("active_key = ? AND status = ?",
		model.ActiveAlertKey(ruleID, serverID), model.EventStatusFiring).First(&ev).Error
	return &ev, err
}

func (r *AlertEventRepository) Create(ev *model.AlertEvent) error {
	return r.db.Create(ev).Error
}

// UpsertFiring 原子创建或刷新活动告警。
// active_key 的唯一索引保证并发评估同一服务器/规则时最多只有一条活动事件。
func (r *AlertEventRepository) UpsertFiring(ev *model.AlertEvent) (*model.AlertEvent, error) {
	key := model.ActiveAlertKey(ev.AlertRuleID, ev.ServerID)
	ev.ActiveKey = &key
	ev.Status = model.EventStatusFiring

	err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "active_key"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"current_value":     ev.CurrentValue,
			"threshold":         ev.Threshold,
			"level":             ev.Level,
			"metric":            ev.Metric,
			"status":            model.EventStatusFiring,
			"last_triggered_at": ev.LastTriggeredAt,
			"recovered_at":      nil,
			"updated_at":        ev.LastTriggeredAt,
		}),
	}).Create(ev).Error
	if err != nil {
		return nil, err
	}

	var stored model.AlertEvent
	if err := r.db.Where("active_key = ?", key).First(&stored).Error; err != nil {
		return nil, err
	}
	return &stored, nil
}

// MarkRecovered 标记恢复
func (r *AlertEventRepository) MarkRecovered(id int64) error {
	return r.db.Model(&model.AlertEvent{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       model.EventStatusRecovered,
			"recovered_at": gorm.Expr("CURRENT_TIMESTAMP"),
			"active_key":   nil,
		}).Error
}

// Acknowledge 记录确认信息，但不改变告警生命周期状态。
func (r *AlertEventRepository) Acknowledge(id, userID int64) error {
	res := r.db.Model(&model.AlertEvent{}).
		Where("id = ? AND status = ?", id, model.EventStatusFiring).
		Updates(map[string]interface{}{
			"acked_at": gorm.Expr("CURRENT_TIMESTAMP"),
			"acked_by": userID,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("仅活动告警可以确认")
	}
	return nil
}

// Close 主动关闭告警并释放活动唯一键；若指标继续超限，下次评估会创建新事件。
func (r *AlertEventRepository) Close(id int64) error {
	return r.db.Model(&model.AlertEvent{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     model.EventStatusClosed,
			"active_key": nil,
		}).Error
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
