package router

import (
	"net/http"

	"go-admin/common/global"
	"go-admin/common/log"

	"github.com/gin-gonic/gin"
	"github.com/go-admin-team/go-admin-core/tools/transfer"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Monitor() {
	var r *gin.Engine
	h := global.Cfg.GetEngine()
	if h == nil {
		h = gin.New()
		global.Cfg.SetEngine(h)
	}
	switch h.(type) {
	case *gin.Engine:
		r = h.(*gin.Engine)
	default:
		log.Fatal("not support other engine")
	}
	//开发环境启动监控指标
	r.GET("/metrics", transfer.Handler(promhttp.Handler()))
	//健康检查
	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
}
