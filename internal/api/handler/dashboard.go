package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gosee/internal/response"
	"gosee/internal/service/dashboard"
)

type DashboardHandler struct {
	svc *dashboard.Service
}

func NewDashboardHandler(svc *dashboard.Service) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) Summary(c *gin.Context) {
	s, err := h.svc.Summary()
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.OK(c, s)
}

func (h *DashboardHandler) TopCPU(c *gin.Context) {
	items, err := h.svc.TopCPU()
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.OK(c, items)
}

func (h *DashboardHandler) TopMemory(c *gin.Context) {
	items, err := h.svc.TopMemory()
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.OK(c, items)
}

func (h *DashboardHandler) TopDisk(c *gin.Context) {
	items, err := h.svc.TopDisk()
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.OK(c, items)
}

func (h *DashboardHandler) RecentAlerts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	items, err := h.svc.RecentAlerts(limit)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.OK(c, items)
}
