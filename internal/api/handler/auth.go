package handler

import (
	"github.com/gin-gonic/gin"

	"gosee/internal/api/middleware"
	"gosee/internal/response"
	"gosee/internal/service/auth"
)

type AuthHandler struct {
	svc *auth.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithHTTP(c, 400, "参数错误: "+err.Error())
		return
	}
	res, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *AuthHandler) Profile(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	info, err := h.svc.Profile(uid)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OK(c, info)
}

type changePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req changePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithHTTP(c, 400, "参数错误: "+err.Error())
		return
	}
	uid := middleware.CurrentUserID(c)
	if err := h.svc.ChangePassword(uid, req.OldPassword, req.NewPassword); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OKMsg(c, "密码修改成功")
}
