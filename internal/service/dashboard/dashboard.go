package dashboard

import (
	"sort"
	"time"

	"gosee/internal/model"
	"gosee/internal/repository"
)

// Service 仪表盘聚合服务
type Service struct {
	serverRepo     *repository.ServerRepository
	metricRepo     *repository.ServerMetricRepository
	alertEventRepo *repository.AlertEventRepository
}

func NewService(
	serverRepo *repository.ServerRepository,
	metricRepo *repository.ServerMetricRepository,
	alertEventRepo *repository.AlertEventRepository,
) *Service {
	return &Service{
		serverRepo:     serverRepo,
		metricRepo:     metricRepo,
		alertEventRepo: alertEventRepo,
	}
}

// Summary 服务器状态汇总
type Summary struct {
	Total    int64 `json:"total"`
	Normal   int64 `json:"normal"`
	Warning  int64 `json:"warning"`
	Critical int64 `json:"critical"`
	Offline  int64 `json:"offline"`
}

func (s *Service) Summary() (*Summary, error) {
	servers, err := s.serverRepo.ListEnabled()
	if err != nil {
		return nil, err
	}
	sum := &Summary{Total: int64(len(servers))}
	for _, sv := range servers {
		switch sv.Status {
		case model.ServerStatusNormal:
			sum.Normal++
		case model.ServerStatusWarning:
			sum.Warning++
		case model.ServerStatusCritical:
			sum.Critical++
		case model.ServerStatusOffline:
			sum.Offline++
		}
	}
	return sum, nil
}

// TopItem Top N 项
type TopItem struct {
	ServerID int64   `json:"server_id"`
	Name     string  `json:"name"`
	Host     string  `json:"host"`
	Value    float64 `json:"value"`
}

func (s *Service) top(field string, limit int) ([]TopItem, error) {
	metrics, err := s.metricRepo.LatestMetricsOfEnabled()
	if err != nil {
		return nil, err
	}
	valueOf := func(m repository.LatestMetricView) float64 {
		switch field {
		case "memory":
			return m.MemoryUsage
		case "disk":
			return m.DiskMaxUsage
		default:
			return m.CPUUsage
		}
	}
	sort.Slice(metrics, func(i, j int) bool {
		return valueOf(metrics[i]) > valueOf(metrics[j])
	})
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	result := make([]TopItem, 0, limit)
	for i, m := range metrics {
		if i >= limit {
			break
		}
		result = append(result, TopItem{ServerID: m.ServerID, Name: m.Name, Host: m.Host, Value: valueOf(m)})
	}
	return result, nil
}

func (s *Service) TopCPU() ([]TopItem, error)     { return s.top("cpu", 5) }
func (s *Service) TopMemory() ([]TopItem, error)  { return s.top("memory", 5) }
func (s *Service) TopDisk() ([]TopItem, error)    { return s.top("disk", 5) }

// RecentAlert 最近告警（映射自 alert_events）
type RecentAlert struct {
	ID          int64     `json:"id"`
	ServerName  string    `json:"server_name"`
	Metric      string    `json:"metric"`
	Value       *float64  `json:"value"`
	Level       string    `json:"level"`
	Status      string    `json:"status"`
	TriggeredAt time.Time `json:"triggered_at"`
}

func (s *Service) RecentAlerts(limit int) ([]RecentAlert, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	events, err := s.alertEventRepo.List(limit)
	if err != nil {
		return nil, err
	}
	result := make([]RecentAlert, 0, len(events))
	for _, e := range events {
		result = append(result, RecentAlert{
			ID:          e.ID,
			ServerName:  e.ServerName,
			Metric:      e.Metric,
			Value:       e.CurrentValue,
			Level:       e.Level,
			Status:      e.Status,
			TriggeredAt: e.LastTriggeredAt,
		})
	}
	return result, nil
}
