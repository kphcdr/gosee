package handler

import (
	"github.com/gin-gonic/gin"

	"gosee/internal/response"
	"gosee/internal/scheduler"
)

// CollectorHandler 采集相关接口
type CollectorHandler struct {
	sched *scheduler.Scheduler
}

func NewCollectorHandler(sched *scheduler.Scheduler) *CollectorHandler {
	return &CollectorHandler{sched: sched}
}

// Run 手动触发一次全量采集（同步执行并返回批次统计）
func (h *CollectorHandler) Run(c *gin.Context) {
	result := h.sched.CollectAll()
	response.OK(c, result)
}
