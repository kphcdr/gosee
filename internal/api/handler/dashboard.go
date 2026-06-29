package handler

import (
	"net/http"
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
	groupID, ok := dashboardGroupID(c)
	if !ok {
		return
	}
	s, err := h.svc.Summary(groupID)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.OK(c, s)
}

func (h *DashboardHandler) TopCPU(c *gin.Context) {
	groupID, ok := dashboardGroupID(c)
	if !ok {
		return
	}
	items, err := h.svc.TopCPU(groupID)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.OK(c, items)
}

func (h *DashboardHandler) TopMemory(c *gin.Context) {
	groupID, ok := dashboardGroupID(c)
	if !ok {
		return
	}
	items, err := h.svc.TopMemory(groupID)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.OK(c, items)
}

func (h *DashboardHandler) TopDisk(c *gin.Context) {
	groupID, ok := dashboardGroupID(c)
	if !ok {
		return
	}
	items, err := h.svc.TopDisk(groupID)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.OK(c, items)
}

func (h *DashboardHandler) RecentAlerts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	groupID, ok := dashboardGroupID(c)
	if !ok {
		return
	}
	items, err := h.svc.RecentAlerts(limit, groupID)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.OK(c, items)
}

func dashboardGroupID(c *gin.Context) (*int64, bool) {
	raw := c.Query("group_id")
	if raw == "" {
		return nil, true
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		response.FailWithHTTP(c, http.StatusBadRequest, "group_id 必须是正整数")
		return nil, false
	}
	return &id, true
}
