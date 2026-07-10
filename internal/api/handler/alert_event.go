package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gosee/internal/api/middleware"
	"gosee/internal/response"
	"gosee/internal/service/alert"
)

type AlertEventHandler struct {
	svc *alert.Service
}

func NewAlertEventHandler(svc *alert.Service) *AlertEventHandler {
	return &AlertEventHandler{svc: svc}
}

func (h *AlertEventHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "200"))
	events, err := h.svc.ListEvents(limit)
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.OK(c, events)
}

func (h *AlertEventHandler) Ack(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, "无效的 ID")
		return
	}
	if err := h.svc.AckEvent(id, middleware.CurrentUserID(c)); err != nil {
		response.Fail(c, "操作失败: "+err.Error())
		return
	}
	response.OKMsg(c, "已确认")
}

func (h *AlertEventHandler) Close(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, "无效的 ID")
		return
	}
	if err := h.svc.CloseEvent(id); err != nil {
		response.Fail(c, "操作失败: "+err.Error())
		return
	}
	response.OKMsg(c, "已关闭")
}
