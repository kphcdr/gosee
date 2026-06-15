package api

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gosee/internal/api/handler"
	"gosee/internal/api/middleware"
	"gosee/internal/response"
	"gosee/web"
)

// Handlers 聚合所有 handler，由 main 装配后传入
type Handlers struct {
	Auth        *handler.AuthHandler
	Server      *handler.ServerHandler
	ServerGroup *handler.ServerGroupHandler
	Collector   *handler.CollectorHandler
	AlertRule           *handler.AlertRuleHandler
	AlertEvent          *handler.AlertEventHandler
	NotificationChannel *handler.NotificationChannelHandler
	Dashboard           *handler.DashboardHandler
}

// SetupRouter 注册全局中间件与所有业务路由
func SetupRouter(h Handlers, jwtSecret string) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// 健康检查（无需鉴权）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		// 无需鉴权：登录
		api.POST("/auth/login", h.Auth.Login)

		// 以下均需 JWT 鉴权
		authed := api.Group("")
		authed.Use(middleware.JWTAuth(jwtSecret))
		{
			authed.GET("/auth/profile", h.Auth.Profile)
			authed.PUT("/auth/password", h.Auth.ChangePassword)

			// 服务器分组
			g := authed.Group("/server-groups")
			{
				g.GET("", h.ServerGroup.List)
				g.POST("", h.ServerGroup.Create)
				g.PUT("/:id", h.ServerGroup.Update)
				g.DELETE("/:id", h.ServerGroup.Delete)
			}

			// 服务器
			s := authed.Group("/servers")
			{
				s.GET("", h.Server.List)
				s.POST("", h.Server.Create)
				s.GET("/:id", h.Server.Get)
				s.PUT("/:id", h.Server.Update)
				s.DELETE("/:id", h.Server.Delete)
				s.POST("/:id/test-ssh", h.Server.TestSSH)
				s.POST("/:id/collect", h.Server.Collect)
				s.GET("/:id/metrics", h.Server.Metrics)
				s.GET("/:id/disks", h.Server.Disks)
			}

			// 手动触发全量采集
			authed.POST("/collect/run", h.Collector.Run)

			// 告警规则
			ar := authed.Group("/alert-rules")
			{
				ar.GET("", h.AlertRule.List)
				ar.POST("", h.AlertRule.Create)
				ar.PUT("/:id", h.AlertRule.Update)
				ar.DELETE("/:id", h.AlertRule.Delete)
				ar.POST("/:id/enable", h.AlertRule.Enable)
				ar.POST("/:id/disable", h.AlertRule.Disable)
			}

			// 告警事件
			ae := authed.Group("/alert-events")
			{
				ae.GET("", h.AlertEvent.List)
				ae.POST("/:id/ack", h.AlertEvent.Ack)
				ae.POST("/:id/close", h.AlertEvent.Close)
			}

			// 通知通道
			nc := authed.Group("/notification-channels")
			{
				nc.GET("", h.NotificationChannel.List)
				nc.POST("", h.NotificationChannel.Create)
				nc.PUT("/:id", h.NotificationChannel.Update)
				nc.DELETE("/:id", h.NotificationChannel.Delete)
				nc.POST("/:id/test", h.NotificationChannel.Test)
			}

			// 仪表盘
			d := authed.Group("/dashboard")
			{
				d.GET("/summary", h.Dashboard.Summary)
				d.GET("/top-cpu", h.Dashboard.TopCPU)
				d.GET("/top-memory", h.Dashboard.TopMemory)
				d.GET("/top-disk", h.Dashboard.TopDisk)
				d.GET("/recent-alerts", h.Dashboard.RecentAlerts)
			}
		}
	}

	// 前端静态资源（go:embed 嵌入 web/dist）：
	// 非 /api 路径优先返回静态文件，找不到回退 index.html（SPA history 模式）
	if distFS, err := fs.Sub(web.Dist, "dist"); err == nil {
		r.NoRoute(func(c *gin.Context) {
			serveSPA(c, distFS)
		})
	}

	return r
}

// serveSPA 托管前端 SPA：/api 前缀返回 404 JSON，其余静态文件优先，否则 index.html 兜底
func serveSPA(c *gin.Context, distFS fs.FS) {
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/api") {
		response.FailWithHTTP(c, http.StatusNotFound, "接口不存在")
		return
	}
	reqPath := strings.TrimPrefix(path, "/")
	if reqPath == "" {
		reqPath = "index.html"
	}
	if data, err := fs.ReadFile(distFS, reqPath); err == nil {
		c.Data(http.StatusOK, contentType(reqPath), data)
		return
	}
	// SPA history 路由兜底
	if data, err := fs.ReadFile(distFS, "index.html"); err == nil {
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		return
	}
	response.FailWithHTTP(c, http.StatusNotFound, "前端资源未找到")
}

func contentType(path string) string {
	switch {
	case strings.HasSuffix(path, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(path, ".js"), strings.HasSuffix(path, ".mjs"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(path, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(path, ".svg"):
		return "image/svg+xml"
	case strings.HasSuffix(path, ".png"):
		return "image/png"
	case strings.HasSuffix(path, ".json"):
		return "application/json"
	}
	return "application/octet-stream"
}
