package util

import (
	"embyhub/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetQueryInt 获取查询参数整数值
func GetQueryInt(c *gin.Context, key string, defaultValue int) int {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, model.Response{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, model.Response{
		Code:    code,
		Message: message,
	})
}

// BadRequestResponse 错误请求
func BadRequestResponse(c *gin.Context, message string) {
	ErrorResponse(c, 400, message)
}

// UnauthorizedResponse 未授权
func UnauthorizedResponse(c *gin.Context, message string) {
	ErrorResponse(c, 401, message)
}

// ForbiddenResponse 禁止访问
func ForbiddenResponse(c *gin.Context, message string) {
	ErrorResponse(c, 403, message)
}

// NotFoundResponse 未找到
func NotFoundResponse(c *gin.Context, message string) {
	ErrorResponse(c, 404, message)
}

// InternalErrorResponse 服务器错误
func InternalErrorResponse(c *gin.Context, message string) {
	ErrorResponse(c, 500, message)
}
