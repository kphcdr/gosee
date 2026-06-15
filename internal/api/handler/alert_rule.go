package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"gosee/internal/model"
	"gosee/internal/response"
	"gosee/internal/service/alert"
)

type AlertRuleHandler struct {
	svc *alert.Service
}

func NewAlertRuleHandler(svc *alert.Service) *AlertRuleHandler {
	return &AlertRuleHandler{svc: svc}
}

func (h *AlertRuleHandler) List(c *gin.Context) {
	rules, err := h.svc.ListRules()
	if err != nil {
		response.Fail(c, "查询失败: "+err.Error())
		return
	}
	response.OK(c, rules)
}

func (h *AlertRuleHandler) Create(c *gin.Context) {
	var rule model.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.FailWithHTTP(c, 400, "参数错误: "+err.Error())
		return
	}
	applyDefaults(&rule)
	rule.ID = 0
	if err := h.svc.CreateRule(&rule); err != nil {
		response.Fail(c, "创建失败: "+err.Error())
		return
	}
	response.OK(c, rule)
}

func (h *AlertRuleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, "无效的 ID")
		return
	}
	var rule model.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.FailWithHTTP(c, 400, "参数错误: "+err.Error())
		return
	}
	applyDefaults(&rule)
	rule.ID = id
	if err := h.svc.UpdateRule(&rule); err != nil {
		response.Fail(c, "更新失败: "+err.Error())
		return
	}
	response.OK(c, rule)
}

func (h *AlertRuleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, "无效的 ID")
		return
	}
	if err := h.svc.DeleteRule(id); err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}
	response.OKMsg(c, "删除成功")
}

func (h *AlertRuleHandler) Enable(c *gin.Context) {
	h.toggle(c, true)
}

func (h *AlertRuleHandler) Disable(c *gin.Context) {
	h.toggle(c, false)
}

func (h *AlertRuleHandler) toggle(c *gin.Context, enabled bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, "无效的 ID")
		return
	}
	if err := h.svc.SetRuleEnabled(id, enabled); err != nil {
		response.Fail(c, "操作失败: "+err.Error())
		return
	}
	response.OKMsg(c, "操作成功")
}

// applyDefaults 补全前端未传的默认值
func applyDefaults(rule *model.AlertRule) {
	if rule.ScopeType == "" {
		rule.ScopeType = model.ScopeTypeGlobal
	}
	if rule.ScopeType == model.ScopeTypeGlobal {
		rule.ScopeID = nil
	}
	if rule.Operator == "" {
		rule.Operator = ">"
	}
	if rule.Level == "" {
		rule.Level = model.AlertLevelWarning
	}
	if rule.DurationTimes <= 0 {
		rule.DurationTimes = 1
	}
	if rule.NotifyIntervalMinutes <= 0 {
		rule.NotifyIntervalMinutes = 60
	}
	if rule.Enabled == 0 {
		// 默认创建即启用（前端可能传 0 表示未设）
		rule.Enabled = 1
	}
}
