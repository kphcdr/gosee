package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gosee/internal/model"
	"gosee/internal/response"
	"gosee/internal/service/notifier"
)

type NotificationChannelHandler struct {
	svc *notifier.Service
}

func NewNotificationChannelHandler(svc *notifier.Service) *NotificationChannelHandler {
	return &NotificationChannelHandler{svc: svc}
}

func (h *NotificationChannelHandler) List(c *gin.Context) {
	chs, err := h.svc.ListChannels()
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.OK(c, chs)
}

func (h *NotificationChannelHandler) Create(c *gin.Context) {
	var ch model.NotificationChannel
	if err := c.ShouldBindJSON(&ch); err != nil {
		response.FailWithHTTP(c, 400, "参数错误: "+err.Error())
		return
	}
	ch.ID = 0
	if err := h.svc.CreateChannel(&ch); err != nil {
		response.Fail(c, "创建失败: "+err.Error())
		return
	}
	response.OK(c, ch)
}

func (h *NotificationChannelHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, "无效的 ID")
		return
	}
	var ch model.NotificationChannel
	if err := c.ShouldBindJSON(&ch); err != nil {
		response.FailWithHTTP(c, 400, "参数错误: "+err.Error())
		return
	}
	ch.ID = id
	if err := h.svc.UpdateChannel(&ch); err != nil {
		response.Fail(c, "更新失败: "+err.Error())
		return
	}
	response.OK(c, ch)
}

func (h *NotificationChannelHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, "无效的 ID")
		return
	}
	if err := h.svc.DeleteChannel(id); err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.OKMsg(c, "删除成功")
}

func (h *NotificationChannelHandler) Test(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, "无效的 ID")
		return
	}
	if err := h.svc.TestChannel(id); err != nil {
		response.Fail(c, "测试发送失败: "+err.Error())
		return
	}
	response.OKMsg(c, "测试消息已发送")
}
