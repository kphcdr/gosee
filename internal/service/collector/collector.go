package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gosee/internal/model"
	"gosee/internal/repository"
	"gosee/internal/sshclient"
)

// SSHResolver 抽象 server.Service 的 SSH 配置解析与状态更新能力，避免循环依赖
type SSHResolver interface {
	ResolveSSHConfig(id int64) (sshclient.Config, error)
	MarkStatus(id int64, status string, lastErr string) error
}

// AlertHook 采集后告警评估钩子（alert.Service 实现），避免 collector 反向依赖 alert 包
type AlertHook interface {
	Evaluate(serverID int64, metric *model.ServerMetric, sshErr error)
}

// Service 采集服务
type Service struct {
	resolver       SSHResolver
	metricRepo     *repository.ServerMetricRepository
	commandTimeout time.Duration
	hook           AlertHook
}

func NewService(resolver SSHResolver, metricRepo *repository.ServerMetricRepository, commandTimeout time.Duration) *Service {
	return &Service{resolver: resolver, metricRepo: metricRepo, commandTimeout: commandTimeout}
}

// SetHook 注入告警评估钩子（采集完成后自动评估）
func (s *Service) SetHook(hook AlertHook) {
	s.hook = hook
}

// notifyHook 触发告警评估（hook 未注入时安全跳过）
func (s *Service) notifyHook(serverID int64, metric *model.ServerMetric, sshErr error) {
	if s.hook != nil {
		s.hook.Evaluate(serverID, metric, sshErr)
	}
}

// CollectResult 单次采集结果
type CollectResult struct {
	ServerID int64               `json:"server_id"`
	Success  bool                `json:"success"`
	Metric   *model.ServerMetric `json:"metric,omitempty"`
	Error    string              `json:"error,omitempty"`
}

// Collect 对单台服务器执行一次采集
func (s *Service) Collect(serverID int64) (*CollectResult, error) {
	cfg, err := s.resolver.ResolveSSHConfig(serverID)
	if err != nil {
		return nil, err
	}

	client, err := sshclient.Connect(cfg)
	if err != nil {
		_ = s.resolver.MarkStatus(serverID, model.ServerStatusOffline, err.Error())
		s.notifyHook(serverID, nil, err) // SSH 失败 → 触发 ssh_fail 规则评估
		return &CollectResult{ServerID: serverID, Error: "SSH 连接失败: " + err.Error()}, nil
	}
	defer client.Close()

	output, err := sshclient.RunCommandWithTimeout(client, CollectScript, s.commandTimeout)
	if err != nil {
		_ = s.resolver.MarkStatus(serverID, model.ServerStatusOffline, "脚本执行失败: "+err.Error())
		s.notifyHook(serverID, nil, err)
		return &CollectResult{ServerID: serverID, Error: "采集脚本执行失败: " + err.Error()}, nil
	}

	data, err := parse(output)
	if err != nil {
		_ = s.resolver.MarkStatus(serverID, model.ServerStatusWarning, "数据解析失败: "+err.Error())
		s.notifyHook(serverID, nil, err)
		return &CollectResult{ServerID: serverID, Error: "采集数据解析失败: " + err.Error()}, nil
	}

	metric, disks := buildMetric(serverID, data, output)
	if err := s.metricRepo.CreateMetric(serverID, metric, disks); err != nil {
		return &CollectResult{ServerID: serverID, Error: "指标入库失败: " + err.Error()}, nil
	}

	_ = s.resolver.MarkStatus(serverID, model.ServerStatusNormal, "")
	s.notifyHook(serverID, metric, nil) // 采集成功 → 评估指标类规则
	return &CollectResult{ServerID: serverID, Success: true, Metric: metric}, nil
}

// Latest 取最新一条指标
func (s *Service) Latest(serverID int64) (*model.ServerMetric, error) {
	return s.metricRepo.Latest(serverID)
}

// Trend 趋势查询
func (s *Service) Trend(serverID int64, since time.Time, limit int) ([]model.ServerMetric, error) {
	return s.metricRepo.Trend(serverID, since, limit)
}

// LatestDisks 最新磁盘明细
func (s *Service) LatestDisks(serverID int64) ([]model.ServerDisk, error) {
	return s.metricRepo.LatestDisks(serverID)
}

var errParseJSON = errors.New("JSON 解析失败")

// parse 解析脚本 JSON 输出
func parse(raw string) (*collectData, error) {
	raw = trimNoise(raw)
	var data collectData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, fmt.Errorf("%w: %v", errParseJSON, err)
	}
	return &data, nil
}

// trimNoise 截取首个 { 到最后一个 }，去除 shell 提示符等噪声
func trimNoise(raw string) string {
	start := strings.IndexByte(raw, '{')
	end := strings.LastIndexByte(raw, '}')
	if start >= 0 && end > start {
		return raw[start : end+1]
	}
	return raw
}

// buildMetric 由采集数据构造入库模型
func buildMetric(serverID int64, data *collectData, raw string) (*model.ServerMetric, []model.ServerDisk) {
	var diskMax float64
	disks := make([]model.ServerDisk, 0, len(data.Disks))
	for _, d := range data.Disks {
		if d.UsagePercent > diskMax {
			diskMax = d.UsagePercent
		}
		disks = append(disks, model.ServerDisk{
			Filesystem:     d.Filesystem,
			MountPoint:     d.MountPoint,
			SizeBytes:      d.SizeBytes,
			UsedBytes:      d.UsedBytes,
			AvailableBytes: d.AvailableBytes,
			UsagePercent:   d.UsagePercent,
		})
	}
	metric := &model.ServerMetric{
		ServerID:          serverID,
		Hostname:          data.Hostname,
		OS:                data.OS,
		CPUUsage:          data.CPU.UsagePercent,
		CPUCores:          data.CPU.Cores,
		MemoryTotalMB:     data.Memory.TotalMB,
		MemoryUsedMB:      data.Memory.UsedMB,
		MemoryAvailableMB: data.Memory.AvailableMB,
		MemoryUsage:       data.Memory.UsagePercent,
		Load1:             data.Load.Load1,
		Load5:             data.Load.Load5,
		Load15:            data.Load.Load15,
		DiskMaxUsage:      diskMax,
		UptimeSeconds:     data.UptimeSeconds,
		RawJSON:           raw,
		CollectedAt:       time.Now(),
	}
	return metric, disks
}
