package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"gosee/internal/repository"
	"gosee/internal/response"
	"gosee/internal/service/collector"
	"gosee/internal/service/server"
)

type ServerHandler struct {
	svc       *server.Service
	collector *collector.Service
}

func NewServerHandler(svc *server.Service, collectorSvc *collector.Service) *ServerHandler {
	return &ServerHandler{svc: svc, collector: collectorSvc}
}

func (h *ServerHandler) List(c *gin.Context) {
	q := repository.ServerListQuery{
		Page:     atoi(c.Query("page")),
		PageSize: atoi(c.Query("page_size")),
		Keyword:  c.Query("keyword"),
	}
	if gid := c.Query("group_id"); gid != "" {
		if v, err := strconv.ParseInt(gid, 10, 64); err == nil {
			q.GroupID = &v
		}
	}
	if en := c.Query("enabled"); en != "" {
		if v, err := strconv.Atoi(en); err == nil {
			v8 := int8(v)
			q.Enabled = &v8
		}
	}
	list, total, err := h.svc.List(q)
	if err != nil {
		response.Fail(c, "查询失败: " + err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": total, "page": q.Page, "page_size": q.PageSize})
}

func (h *ServerHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	s, err := h.svc.Get(id)
	if err != nil {
		response.Fail(c, "服务器不存在")
		return
	}
	response.OK(c, s)
}

func (h *ServerHandler) Create(c *gin.Context) {
	var in server.SaveInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithHTTP(c, 400, "参数错误: " + err.Error())
		return
	}
	s, err := h.svc.Create(in)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, s)
}

func (h *ServerHandler) Update(c *gin.Context) {
	var in server.SaveInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailWithHTTP(c, 400, "参数错误: " + err.Error())
		return
	}
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	in.ID = id
	s, err := h.svc.Update(in)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, s)
}

func (h *ServerHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.Delete(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OKMsg(c, "删除成功")
}

func (h *ServerHandler) TestSSH(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.svc.TestSSH(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OKMsg(c, "SSH 连接正常")
}

// Collect 手动触发一次采集
func (h *ServerHandler) Collect(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	res, err := h.collector.Collect(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	if !res.Success {
		response.Fail(c, res.Error)
		return
	}
	response.OK(c, res)
}

// Metrics 指标趋势，默认最近 24 小时、最多 200 条
func (h *ServerHandler) Metrics(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	hours := atoi(c.DefaultQuery("hours", "24"))
	if hours <= 0 {
		hours = 24
	}
	limit := atoi(c.DefaultQuery("limit", "200"))
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	list, err := h.collector.Trend(id, since, limit)
	if err != nil {
		response.Fail(c, "查询失败: " + err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

// Disks 最新一次采集的磁盘明细
func (h *ServerHandler) Disks(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	disks, err := h.collector.LatestDisks(id)
	if err != nil {
		response.Fail(c, "查询失败: " + err.Error())
		return
	}
	response.OK(c, disks)
}

// atoi 安全字符串转整数，空串或异常返回 0
func atoi(s string) int {
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}
