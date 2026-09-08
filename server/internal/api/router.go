package api

// router.go 路由注册（全部静态路径，参数走 query / body）

import (
	"github.com/gin-gonic/gin"

	"workbench/server/internal/auth"
	"workbench/server/internal/store"
)

func NewRouter(st *store.Store, authSvc *auth.Service) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// 免鉴权
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"code": 0, "msg": "ok", "data": "ok"})
	})
	r.POST("/api/auth/login", authSvc.Login)
	r.POST("/api/auth/change-password", authSvc.Middleware(), authSvc.ChangePassword)

	// JWT 保护组
	api := r.Group("/api", authSvc.Middleware())
	h := &handlers{st: st}
	{
		api.GET("/news/list", h.newsList)
		api.POST("/news/mark-read", h.newsMarkRead)
		api.GET("/debt/list", h.debtList)
		api.POST("/debt/mark-read", h.debtMarkRead)
		api.GET("/house/list", h.houseList)
		api.POST("/house/mark-read", h.houseMarkRead)

		api.GET("/favorite/list", h.favoriteList)
		api.POST("/favorite/add", h.favoriteAdd)
		api.POST("/favorite/remove", h.favoriteRemove)

		api.GET("/stats/overview", h.statsOverview)

		api.GET("/source/list", h.sourceList)
		api.POST("/source/toggle", h.sourceToggle)
		api.POST("/source/run", h.sourceRun)

		api.GET("/config/list", h.configList)
		api.POST("/config/update", h.configUpdate)

		api.GET("/backup/list", h.backupList)
		api.POST("/backup/run", h.backupRun)
	}
	return r
}
