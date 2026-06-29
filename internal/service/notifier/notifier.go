package notifier

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"gosee/internal/model"
	"gosee/internal/repository"
	"gosee/internal/utils"
)

// Service 通知服务：飞书 webhook 发送 + 防重复 + 发送记录
type Service struct {
	channelRepo *repository.NotificationChannelRepository
	notifyRepo  *repository.AlertNotificationRepository
	eventRepo   *repository.AlertEventRepository
	serverRepo  *repository.ServerRepository
	ruleRepo    *repository.AlertRuleRepository
}

func NewService(
	channelRepo *repository.NotificationChannelRepository,
	notifyRepo *repository.AlertNotificationRepository,
	eventRepo *repository.AlertEventRepository,
	serverRepo *repository.ServerRepository,
	ruleRepo *repository.AlertRuleRepository,
) *Service {
	return &Service{
		channelRepo: channelRepo,
		notifyRepo:  notifyRepo,
		eventRepo:   eventRepo,
		serverRepo:  serverRepo,
		ruleRepo:    ruleRepo,
	}
}

// NotifyOnFired 告警触发/持续通知。实现 alert.NotifierHook。
// 防重复：距上次通知不足 rule.notify_interval 则跳过。
func (s *Service) NotifyOnFired(event *model.AlertEvent) {
	if event == nil {
		return
	}
	rule, err := s.ruleRepo.FindByID(event.AlertRuleID)
	interval := time.Duration(60) * time.Minute
	if err == nil && rule.NotifyIntervalMinutes > 0 {
		interval = time.Duration(rule.NotifyIntervalMinutes) * time.Minute
	}
	if event.LastNotifiedAt != nil && time.Since(*event.LastNotifiedAt) < interval {
		return // 间隔内，跳过（PRD 6.5）
	}
	server, err := s.serverRepo.FindByID(event.ServerID)
	if err != nil {
		return
	}
	text := s.buildAlertText(server, event)
	s.sendToAll(event, text)
	_ = s.eventRepo.UpdateNotified(event.ID) // 更新 last_notified_at + notify_count
}

// NotifyOnRecovered 恢复通知。实现 alert.NotifierHook。
func (s *Service) NotifyOnRecovered(event *model.AlertEvent) {
	if event == nil {
		return
	}
	server, err := s.serverRepo.FindByID(event.ServerID)
	if err != nil {
		return
	}
	text := s.buildRecoverText(server, event)
	s.sendToAll(event, text)
}

// sendToAll 向所有启用通道发送
func (s *Service) sendToAll(event *model.AlertEvent, text string) {
	channels, err := s.channelRepo.ListEnabled()
	if err != nil {
		return
	}
	for i := range channels {
		s.sendToOne(&channels[i], event, text)
	}
}

func (s *Service) sendToOne(ch *model.NotificationChannel, event *model.AlertEvent, text string) {
	now := time.Now()
	var status, response string

	switch ch.Type {
	case model.ChannelTypeFeishu:
		cfg, err := parseFeishuConfig(ch.Config)
		if err != nil {
			s.record(event.ID, ch.ID, model.NotifyStatusFailed, "config 解析失败: "+err.Error(), nil)
			return
		}
		resp, err := sendFeishu(cfg.Webhook, cfg.Secret, text)
		if err != nil {
			status = model.NotifyStatusFailed
			response = err.Error()
		} else {
			status = model.NotifyStatusSuccess
			response = resp
		}
	default:
		// telegram/email 第一版未实现
		s.record(event.ID, ch.ID, model.NotifyStatusFailed, "通道类型暂未实现: "+ch.Type, nil)
		return
	}

	if utils.Logger != nil {
		if status == model.NotifyStatusSuccess {
			utils.Logger.Info("通知已发送",
				zap.String("channel", ch.Name),
				zap.Int64("event", event.ID),
			)
		} else {
			utils.Logger.Warn("通知发送失败",
				zap.String("channel", ch.Name),
				zap.Int64("event", event.ID),
				zap.String("response", response),
			)
		}
	}
	s.record(event.ID, ch.ID, status, response, &now)
}

func (s *Service) record(eventID, channelID int64, status, response string, sentAt *time.Time) {
	_ = s.notifyRepo.Create(&model.AlertNotification{
		AlertEventID: eventID,
		ChannelID:    channelID,
		Status:       status,
		Response:     response,
		SentAt:       sentAt,
	})
}

// SendTest 测试发送（handler /test 调用），返回错误供前端提示
func (s *Service) SendTest(ch *model.NotificationChannel) error {
	switch ch.Type {
	case model.ChannelTypeFeishu:
		cfg, err := parseFeishuConfig(ch.Config)
		if err != nil {
			return fmt.Errorf("config 解析失败: %w", err)
		}
		text := fmt.Sprintf("【gosee 测试通知】\n\n通道：%s\n时间：%s", ch.Name, formatTime(time.Now()))
		_, err = sendFeishu(cfg.Webhook, cfg.Secret, text)
		return err
	default:
		return fmt.Errorf("通道类型 %s 暂未实现", ch.Type)
	}
}

// ===== 通道 CRUD（handler 用）=====

func (s *Service) ListChannels() ([]model.NotificationChannel, error) {
	return s.channelRepo.List()
}

func (s *Service) CreateChannel(ch *model.NotificationChannel) error {
	return s.channelRepo.Create(ch)
}

func (s *Service) UpdateChannel(ch *model.NotificationChannel) error {
	return s.channelRepo.Update(ch)
}

func (s *Service) DeleteChannel(id int64) error {
	return s.channelRepo.Delete(id)
}

// TestChannel 测试发送指定通道
func (s *Service) TestChannel(id int64) error {
	ch, err := s.channelRepo.FindByID(id)
	if err != nil {
		return err
	}
	return s.SendTest(ch)
}

// ===== 文案（PRD 16）=====

func (s *Service) buildAlertText(server *model.Server, event *model.AlertEvent) string {
	if event.Metric == model.MetricSSHFail {
		return fmt.Sprintf(`【服务器离线】

服务器：%s
IP：%s
原因：SSH 连接失败
连续失败：%s 次
时间：%s`,
			server.Name, server.Host, formatCount(event.CurrentValue), formatTime(event.LastTriggeredAt))
	}
	return fmt.Sprintf(`【服务器告警】

服务器：%s
IP：%s
指标：%s
当前值：%s
阈值：%s
级别：%s
时间：%s`,
		server.Name, server.Host,
		metricLabel(event.Metric),
		formatValue(event.Metric, event.CurrentValue),
		formatValue(event.Metric, event.Threshold),
		event.Level,
		formatTime(event.LastTriggeredAt),
	)
}

func (s *Service) buildRecoverText(server *model.Server, event *model.AlertEvent) string {
	recovered := "现在"
	if event.RecoveredAt != nil {
		recovered = formatTime(*event.RecoveredAt)
	}
	return fmt.Sprintf(`【服务器恢复】

服务器：%s
IP：%s
指标：%s
当前值：%s
恢复时间：%s`,
		server.Name, server.Host,
		metricLabel(event.Metric),
		formatValue(event.Metric, event.CurrentValue),
		recovered,
	)
}

// ===== 飞书 webhook =====

type feishuConfig struct {
	Webhook string `json:"webhook"`
	Secret  string `json:"secret"`
}

func parseFeishuConfig(raw json.RawMessage) (feishuConfig, error) {
	var c feishuConfig
	err := json.Unmarshal(raw, &c)
	return c, err
}

func sendFeishu(webhook, secret, text string) (string, error) {
	if webhook == "" {
		return "", fmt.Errorf("webhook 未配置")
	}
	payload := map[string]interface{}{
		"msg_type": "text",
		"content":  map[string]string{"text": text},
	}
	if secret != "" {
		ts := strconv.FormatInt(time.Now().Unix(), 10)
		payload["timestamp"] = ts
		payload["sign"] = feishuSign(ts, secret)
	}
	b, _ := json.Marshal(payload)
	resp, err := http.Post(webhook, "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

// feishuSign 飞书自定义机器人签名：HMAC-SHA256(key=timestamp+"\n"+secret) → base64
func feishuSign(timestamp, secret string) string {
	stringToSign := timestamp + "\n" + secret
	h := hmac.New(sha256.New, []byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// ===== 辅助 =====

func metricLabel(metric string) string {
	switch metric {
	case model.MetricCPUUsage:
		return "CPU 使用率"
	case model.MetricMemoryUsage:
		return "内存使用率"
	case model.MetricDiskUsage:
		return "磁盘使用率"
	case model.MetricLoad1:
		return "1 分钟负载"
	case model.MetricLoad5:
		return "5 分钟负载"
	case model.MetricLoad15:
		return "15 分钟负载"
	case model.MetricSSHFail:
		return "SSH 连接"
	}
	return metric
}

func formatValue(metric string, v *float64) string {
	if v == nil {
		return "-"
	}
	switch metric {
	case model.MetricCPUUsage, model.MetricMemoryUsage, model.MetricDiskUsage:
		return fmt.Sprintf("%.2f%%", *v)
	case model.MetricSSHFail:
		return fmt.Sprintf("%.0f 次", *v)
	}
	return fmt.Sprintf("%.2f", *v)
}

func formatCount(v *float64) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%.0f", *v)
}

func formatTime(t time.Time) string {
	return t.In(time.Local).Format("2006-01-02 15:04:05")
}
