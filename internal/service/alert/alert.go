package alert

import (
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"gosee/internal/model"
	"gosee/internal/repository"
	"gosee/internal/utils"
)

// NotifierHook 通知钩子（notifier.Service 实现），告警触发/恢复后异步调用
type NotifierHook interface {
	NotifyOnFired(event *model.AlertEvent)
	NotifyOnRecovered(event *model.AlertEvent)
}

// Service 告警服务：规则 CRUD + 事件查询 + 评估器
type Service struct {
	ruleRepo     *repository.AlertRuleRepository
	eventRepo    *repository.AlertEventRepository
	metricRepo   *repository.ServerMetricRepository
	serverRepo   *repository.ServerRepository
	notifierHook NotifierHook
}

func NewService(
	ruleRepo *repository.AlertRuleRepository,
	eventRepo *repository.AlertEventRepository,
	metricRepo *repository.ServerMetricRepository,
	serverRepo *repository.ServerRepository,
) *Service {
	return &Service{
		ruleRepo:   ruleRepo,
		eventRepo:  eventRepo,
		metricRepo: metricRepo,
		serverRepo: serverRepo,
	}
}

// SetNotifierHook 注入通知钩子（告警触发/恢复后发送通知）
func (s *Service) SetNotifierHook(hook NotifierHook) {
	s.notifierHook = hook
}

// notifyFired 异步触发告警通知（不阻塞采集流程）
func (s *Service) notifyFired(event *model.AlertEvent) {
	if s.notifierHook != nil {
		go s.notifierHook.NotifyOnFired(event)
	}
}

func (s *Service) notifyRecovered(event *model.AlertEvent) {
	if s.notifierHook != nil {
		go s.notifierHook.NotifyOnRecovered(event)
	}
}

// ===== 规则 CRUD（handler 用）=====

func (s *Service) ListRules() ([]model.AlertRule, error) {
	return s.ruleRepo.List()
}

func (s *Service) CreateRule(rule *model.AlertRule) error {
	return s.ruleRepo.Create(rule)
}

func (s *Service) UpdateRule(rule *model.AlertRule) error {
	return s.ruleRepo.Update(rule)
}

func (s *Service) DeleteRule(id int64) error {
	return s.ruleRepo.Delete(id)
}

func (s *Service) SetRuleEnabled(id int64, enabled bool) error {
	v := int8(0)
	if enabled {
		v = 1
	}
	return s.ruleRepo.SetEnabled(id, v)
}

// ===== 事件（handler 用）=====

func (s *Service) ListEvents(limit int) ([]repository.AlertEventView, error) {
	return s.eventRepo.List(limit)
}

func (s *Service) AckEvent(id int64) error {
	return s.eventRepo.UpdateStatus(id, model.EventStatusAcked)
}

func (s *Service) CloseEvent(id int64) error {
	return s.eventRepo.UpdateStatus(id, model.EventStatusClosed)
}

// ===== 评估器（collector 采集后调用）=====

// Evaluate 采集后评估告警。metric 非 nil 表示采集成功；sshErr 非 nil 表示 SSH 失败。
func (s *Service) Evaluate(serverID int64, metric *model.ServerMetric, sshErr error) {
	server, err := s.serverRepo.FindByID(serverID)
	if err != nil {
		return
	}
	rules, err := s.ruleRepo.ApplicableRules(serverID, server.GroupID)
	if err != nil {
		return
	}
	sshFailureCount := 0
	if sshErr != nil {
		sshFailureCount, err = s.serverRepo.IncrementSSHFailureCount(serverID)
	} else {
		err = s.serverRepo.ResetSSHFailureCount(serverID)
	}
	if err != nil {
		return
	}
	for i := range rules {
		s.evaluateRule(server, &rules[i], metric, sshErr, sshFailureCount)
	}
}

func (s *Service) evaluateRule(server *model.Server, rule *model.AlertRule, metric *model.ServerMetric, sshErr error, sshFailureCount int) {
	// SSH 规则同样严格遵循运算符、阈值和连续次数。
	if rule.Metric == model.MetricSSHFail {
		if sshErr != nil && sshFailureCount >= rule.DurationTimes && compare(float64(sshFailureCount), rule.Operator, rule.Threshold) {
			s.fireEvent(server, rule, float64(sshFailureCount), rule.Threshold)
		} else {
			if sshErr == nil {
				s.recoverEvent(server, rule)
			}
		}
		return
	}
	if metric == nil {
		return
	}
	value, ok := metricValue(metric, rule.Metric)
	if !ok {
		return
	}
	threshold := rule.Threshold
	// load 类规则阈值=0 表示动态阈值（cpu_cores * 2，PRD 6.4）
	if isLoadMetric(rule.Metric) && rule.Threshold == 0 {
		threshold = float64(metric.CPUCores) * 2
	}
	// 连续 N 次：读最近 N 条 metric，全部超阈值才触发
	recent, _ := s.metricRepo.RecentN(server.ID, rule.DurationTimes)
	if compare(value, rule.Operator, threshold) && allExceed(recent, rule) {
		s.fireEvent(server, rule, value, threshold)
	} else {
		s.recoverEvent(server, rule)
	}
}

func (s *Service) fireEvent(server *model.Server, rule *model.AlertRule, value, threshold float64) {
	existing, err := s.eventRepo.FindFiring(rule.ID, server.ID)
	if err == nil && existing != nil {
		v := value
		if err := s.eventRepo.TouchFiring(existing.ID, &v); err == nil && utils.Logger != nil {
			utils.Logger.Info("告警持续",
				zap.String("server", server.Name),
				zap.String("rule", rule.Name),
				zap.Float64("value", value),
			)
		}
		existing.CurrentValue = &v
		existing.LastTriggeredAt = time.Now()
		s.notifyFired(existing)
		return
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	v := value
	th := threshold
	now := time.Now()
	ev := &model.AlertEvent{
		ServerID:         server.ID,
		AlertRuleID:      rule.ID,
		Metric:           rule.Metric,
		CurrentValue:     &v,
		Threshold:        &th,
		Level:            rule.Level,
		Status:           model.EventStatusFiring,
		FirstTriggeredAt: now,
		LastTriggeredAt:  now,
	}
	if err := s.eventRepo.Create(ev); err == nil && utils.Logger != nil {
		utils.Logger.Info("告警触发",
			zap.String("server", server.Name),
			zap.String("rule", rule.Name),
			zap.String("level", rule.Level),
			zap.Float64("value", value),
			zap.Float64("threshold", threshold),
		)
		s.notifyFired(ev)
	}
}

func (s *Service) recoverEvent(server *model.Server, rule *model.AlertRule) {
	existing, err := s.eventRepo.FindFiring(rule.ID, server.ID)
	if err != nil || existing == nil {
		return
	}
	if err := s.eventRepo.MarkRecovered(existing.ID); err == nil && utils.Logger != nil {
		utils.Logger.Info("告警恢复",
			zap.String("server", server.Name),
			zap.String("rule", rule.Name),
		)
		s.notifyRecovered(existing)
	}
}

// ===== 辅助函数 =====

func metricValue(m *model.ServerMetric, metric string) (float64, bool) {
	switch metric {
	case model.MetricCPUUsage:
		return m.CPUUsage, true
	case model.MetricMemoryUsage:
		return m.MemoryUsage, true
	case model.MetricDiskUsage:
		return m.DiskMaxUsage, true
	case model.MetricLoad1:
		return m.Load1, true
	case model.MetricLoad5:
		return m.Load5, true
	case model.MetricLoad15:
		return m.Load15, true
	}
	return 0, false
}

func isLoadMetric(metric string) bool {
	return metric == model.MetricLoad1 || metric == model.MetricLoad5 || metric == model.MetricLoad15
}

func compare(value float64, op string, threshold float64) bool {
	switch op {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	}
	return false
}

// allExceed 最近 N 条是否全部超阈值（含 load 动态阈值）
func allExceed(recent []model.ServerMetric, rule *model.AlertRule) bool {
	if len(recent) == 0 {
		return false
	}
	for i := range recent {
		value, ok := metricValue(&recent[i], rule.Metric)
		if !ok {
			return false
		}
		threshold := rule.Threshold
		if isLoadMetric(rule.Metric) && rule.Threshold == 0 {
			threshold = float64(recent[i].CPUCores) * 2
		}
		if !compare(value, rule.Operator, threshold) {
			return false
		}
	}
	return true
}
