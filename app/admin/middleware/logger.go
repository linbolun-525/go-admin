package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"go-admin/app/admin/models"
	"go-admin/app/admin/models/system"
	"go-admin/app/admin/service"
	"go-admin/common/log"
	"go-admin/tools"
	"go-admin/tools/config"
)

// LoggerToFile 日志记录到文件
func LoggerToFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()
		// 处理请求
		c.Next()

		bd, bl := c.Get("body")
		var body = ""
		if bl {
			body = bd.(string)
		}

		rt, bl := c.Get("result")
		var result = ""
		if bl {
			result = rt.(string)
		}

		st, bl := c.Get("status")
		var statusBus = 0
		if bl {
			statusBus = st.(int)
		}

		// 请求方式
		reqMethod := c.Request.Method
		// 请求路由
		reqUri := c.Request.RequestURI
		// 状态码
		statusCode := c.Writer.Status()
		// 请求IP
		clientIP := c.ClientIP()
		// 结束时间
		endTime := time.Now()
		// 执行时间
		latencyTime := endTime.Sub(startTime)
		// 日志格式
		logData := map[string]interface{}{
			"statusCode":  statusCode,
			"latencyTime": latencyTime,
			"clientIP":    clientIP,
			"method":      reqMethod,
			"uri":         reqUri,
		}
		log.Info(logData)
		if c.Request.Method != "GET" && c.Request.Method != "OPTIONS" && config.LoggerConfig.EnabledDB {
			SetDBOperLog(c, clientIP, statusCode, reqUri, reqMethod, latencyTime, body, result, statusBus)
		}
	}
}

// SetDBOperLog 写入操作日志表 fixme 该方法后续即将弃用
func SetDBOperLog(c *gin.Context, clientIP string, statusCode int, reqUri string, reqMethod string, latencyTime time.Duration, body string, result string, status int) {
	menu := models.Menu{}
	menu.Path = reqUri
	menu.Action = reqMethod
	menuList, _ := menu.Get()
	sysOperaLog := system.SysOperaLog{}
	sysOperaLog.OperIp = clientIP
	sysOperaLog.OperLocation = tools.GetLocation(clientIP)
	sysOperaLog.Status = tools.IntToString(statusCode)
	sysOperaLog.OperName = tools.GetUserName(c)
	sysOperaLog.RequestMethod = c.Request.Method
	sysOperaLog.OperUrl = reqUri
	sysOperaLog.OperParam = body
	if reqUri == "/login" {
		sysOperaLog.BusinessType = "10"
		sysOperaLog.Title = "用户登录"
		sysOperaLog.OperName = "-"
	} else if strings.Contains(reqUri, "/api/v1/logout") {
		sysOperaLog.BusinessType = "11"
	} else if strings.Contains(reqUri, "/api/v1/getCaptcha") {
		sysOperaLog.BusinessType = "12"
		sysOperaLog.Title = "验证码"
	} else {
		if reqMethod == "POST" {
			sysOperaLog.BusinessType = "1"
		} else if reqMethod == "PUT" {
			sysOperaLog.BusinessType = "2"
		} else if reqMethod == "DELETE" {
			sysOperaLog.BusinessType = "3"
		}
	}
	sysOperaLog.Method = reqMethod
	if len(menuList) > 0 {
		sysOperaLog.Title = menuList[0].Title
	}
	sysOperaLog.CreateBy = tools.GetUserId(c)
	sysOperaLog.OperTime = tools.GetCurrentTime()
	sysOperaLog.LatencyTime = fmt.Sprintf("%v", latencyTime)

	sysOperaLog.JsonResult = result
	sysOperaLog.UserAgent = c.Request.UserAgent()
	if status == 200 {
		sysOperaLog.Status = "2"
	} else {
		sysOperaLog.Status = "1"
	}
	msgID := tools.GenerateMsgIDFromContext(c)
	db, err := tools.GetOrm(c)
	if err != nil {
		log.Errorf("msgID[%s] 获取Orm失败, error:%s", msgID, err)
	}
	serviceOperaLog := service.SysOperaLog{}
	serviceOperaLog.Orm = db
	_ = serviceOperaLog.InsertSysOperaLog(&sysOperaLog)
}
