package database

import (
	"fmt"

	"gorm.io/gorm"

	"gosee/internal/model"
)

// migrateAlertEventState 把旧版 acked 生命周期状态迁移为独立确认信息，
// 并为每个服务器/规则保留唯一一条活动告警。
func migrateAlertEventState(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.AlertEvent{}).
			Where("status = ?", "acked").
			Updates(map[string]interface{}{
				"status":   model.EventStatusFiring,
				"acked_at": gorm.Expr("COALESCE(acked_at, updated_at)"),
			}).Error; err != nil {
			return fmt.Errorf("迁移已确认告警状态失败: %w", err)
		}

		var active []model.AlertEvent
		if err := tx.Where("status = ?", model.EventStatusFiring).
			Order("alert_rule_id ASC, server_id ASC, last_triggered_at DESC, id DESC").
			Find(&active).Error; err != nil {
			return fmt.Errorf("查询活动告警失败: %w", err)
		}

		groups := make(map[string][]model.AlertEvent)
		for i := range active {
			key := model.ActiveAlertKey(active[i].AlertRuleID, active[i].ServerID)
			groups[key] = append(groups[key], active[i])
		}

		for key, events := range groups {
			if len(events) == 1 && events[0].ActiveKey != nil && *events[0].ActiveKey == key {
				continue
			}
			ids := make([]int64, 0, len(events))
			for i := range events {
				ids = append(ids, events[i].ID)
			}
			// 先清空旧键，确保迁移可重复执行；排序后的第一条是最近活动事件。
			if err := tx.Model(&model.AlertEvent{}).Where("id IN ?", ids).
				Update("active_key", nil).Error; err != nil {
				return fmt.Errorf("重置活动告警键失败: %w", err)
			}
			if len(ids) > 1 {
				if err := tx.Model(&model.AlertEvent{}).Where("id IN ?", ids[1:]).
					Updates(map[string]interface{}{
						"status":     model.EventStatusClosed,
						"active_key": nil,
					}).Error; err != nil {
					return fmt.Errorf("合并重复活动告警失败: %w", err)
				}
			}
			if err := tx.Model(&model.AlertEvent{}).Where("id = ?", ids[0]).
				Update("active_key", key).Error; err != nil {
				return fmt.Errorf("设置活动告警键失败: %w", err)
			}
		}
		return nil
	})
}
