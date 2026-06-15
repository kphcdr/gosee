package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gosee/internal/response"
	"gosee/internal/service/server_group"
)

type ServerGroupHandler struct {
	svc *server_group.Service
}

func NewServerGroupHandler(svc *server_group.Service) *ServerGroupHandler {
	return &ServerGroupHandler{svc: svc}
}

func (h *ServerGroupHandler) List(c *gin.Context) {
	groups, err := h.svc.List(c.Query("keyword"))
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.OK(c, groups)
}

func (h *ServerGroupHandler) Create(c *gin.Context) {
	var in server_group.SaveInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithHTTP(c, 400, "参数错误: "+err.Error())
		return
	}
	g, err := h.svc.Create(in)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, g)
}

func (h *ServerGroupHandler) Update(c *gin.Context) {
	var in server_group.SaveInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithHTTP(c, 400, "参数错误: "+err.Error())
		return
	}
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	in.ID = id
	g, err := h.svc.Update(in)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, g)
}

func (h *ServerGroupHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OKMsg(c, "删除成功")
}
