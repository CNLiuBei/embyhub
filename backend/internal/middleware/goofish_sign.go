// Package middleware 闲管家签名验证中间件
package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"feiniu-user-system/internal/models"
	"feiniu-user-system/internal/service"

	"github.com/gin-gonic/gin"
)

// GoofishSignMiddleware 闲管家签名验证中间件
type GoofishSignMiddleware struct {
	service *service.GoofishService
}

// NewGoofishSignMiddleware 创建签名验证中间件
func NewGoofishSignMiddleware(svc *service.GoofishService) *GoofishSignMiddleware {
	return &GoofishSignMiddleware{service: svc}
}

// Verify 验证签名中间件
func (m *GoofishSignMiddleware) Verify() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 从URL参数获取签名信息
		mchID := c.Query("mch_id")
		timestamp := c.Query("timestamp")
		sign := c.Query("sign")

		// 读取请求body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			m.logAndRespond(c, startTime, http.StatusBadRequest, 500, "读取请求体失败", string(bodyBytes))
			return
		}
		// 重新设置body供后续handler使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		body := string(bodyBytes)
		if body == "" {
			body = "{}"
		}

		// 验证签名
		code, err := m.service.VerifySign(mchID, timestamp, sign, body)
		if err != nil {
			m.logAndRespond(c, startTime, http.StatusOK, code, err.Error(), body)
			return
		}

		// 将body存入context供后续使用
		c.Set("request_body", body)
		c.Set("start_time", startTime)

		c.Next()
	}
}

// logAndRespond 记录日志并响应
func (m *GoofishSignMiddleware) logAndRespond(c *gin.Context, startTime time.Time, httpStatus, code int, msg, requestBody string) {
	duration := time.Since(startTime).Milliseconds()

	// 记录日志
	log := &models.GoofishLog{
		Endpoint:     c.Request.URL.Path,
		Method:       c.Request.Method,
		RequestBody:  requestBody,
		ResponseBody: `{"code":` + string(rune(code)) + `,"msg":"` + msg + `"}`,
		ResponseCode: code,
		Duration:     duration,
		ClientIP:     c.ClientIP(),
	}
	m.service.LogAPICall(log)

	// 响应
	c.JSON(httpStatus, gin.H{
		"code": code,
		"msg":  msg,
	})
	c.Abort()
}

// LogResponse 记录响应日志中间件
func (m *GoofishSignMiddleware) LogResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取开始时间
		startTime, exists := c.Get("start_time")
		if !exists {
			startTime = time.Now()
		}

		// 获取请求body
		requestBody, _ := c.Get("request_body")
		reqBodyStr, _ := requestBody.(string)

		// 使用自定义ResponseWriter捕获响应
		rw := &responseWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = rw

		c.Next()

		// 记录日志
		duration := time.Since(startTime.(time.Time)).Milliseconds()
		log := &models.GoofishLog{
			Endpoint:     c.Request.URL.Path,
			Method:       c.Request.Method,
			RequestBody:  reqBodyStr,
			ResponseBody: rw.body.String(),
			ResponseCode: c.Writer.Status(),
			Duration:     duration,
			ClientIP:     c.ClientIP(),
		}
		m.service.LogAPICall(log)
	}
}

// responseWriter 自定义ResponseWriter用于捕获响应内容
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
