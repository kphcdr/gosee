package scheduler

import (
	"sync"
	"sync/atomic"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"gosee/internal/config"
	"gosee/internal/model"
	"gosee/internal/repository"
	"gosee/internal/service/collector"
	"gosee/internal/utils"
)

// BatchResult 一次批量采集的统计
type BatchResult struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

// Scheduler 定时采集调度器
type Scheduler struct {
	cron        *cron.Cron
	collector   *collector.Service
	serverRepo  *repository.ServerRepository
	workerCount int
	running     int32 // atomic，防止任务重叠
}

func New(svc *collector.Service, serverRepo *repository.ServerRepository, cfg *config.CollectorConfig) *Scheduler {
	workers := cfg.WorkerCount
	if workers <= 0 {
		workers = 10
	}
	return &Scheduler{
		cron:        cron.New(),
		collector:   svc,
		serverRepo:  serverRepo,
		workerCount: workers,
	}
}

// Start 注册定时任务并启动。interval 形如 "10m"，内部转 @every 10m。
func (s *Scheduler) Start(interval string) error {
	spec := "@every " + interval
	if _, err := s.cron.AddFunc(spec, s.safeCollectAll); err != nil {
		return err
	}
	s.cron.Start()
	utils.Logger.Info("定时采集已启动", zap.String("spec", spec), zap.Int("workers", s.workerCount))
	return nil
}

// Stop 优雅停止调度
func (s *Scheduler) Stop() {
	if s.cron != nil {
		ctx := s.cron.Stop()
		<-ctx.Done()
		utils.Logger.Info("定时采集已停止")
	}
}

// safeCollectAll 带防重叠保护的定时任务入口
func (s *Scheduler) safeCollectAll() {
	if !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		utils.Logger.Warn("上一轮采集仍在进行，跳过本次触发")
		return
	}
	defer atomic.StoreInt32(&s.running, 0)
	s.CollectAll()
}

// CollectAll 全量采集所有启用服务器（Worker Pool 并发）。
// 同步执行完毕并返回统计；既被定时器调用，也供手动触发接口使用。
func (s *Scheduler) CollectAll() BatchResult {
	servers, err := s.serverRepo.ListEnabled()
	if err != nil {
		utils.Logger.Error("查询启用服务器失败", zap.Error(err))
		return BatchResult{}
	}
	if len(servers) == 0 {
		utils.Logger.Info("无启用的服务器，跳过采集")
		return BatchResult{}
	}

	taskCh := make(chan model.Server)
	resultCh := make(chan bool, len(servers))

	workers := s.workerCount
	if workers > len(servers) {
		workers = len(servers)
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for srv := range taskCh {
				res, err := s.collector.Collect(srv.ID)
				resultCh <- (err == nil && res != nil && res.Success)
			}
		}()
	}

	for _, srv := range servers {
		taskCh <- srv
	}
	close(taskCh)

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	success, failed := 0, 0
	for ok := range resultCh {
		if ok {
			success++
		} else {
			failed++
		}
	}

	result := BatchResult{Total: len(servers), Success: success, Failed: failed}
	utils.Logger.Info("采集批次完成",
		zap.Int("total", result.Total),
		zap.Int("success", result.Success),
		zap.Int("failed", result.Failed),
	)
	return result
}
