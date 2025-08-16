package utils

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// APIResponse represents the standard API response format
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse sends a success response
func SuccessResponse(r *ghttp.Request, data interface{}) {
	r.Response.WriteJsonExit(APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// ErrorResponse sends an error response
func ErrorResponse(r *ghttp.Request, code int, message string) {
	r.Response.WriteJsonExit(APIResponse{
		Code:    code,
		Message: message,
	})
}

// HealthResponse sends a health check response
func HealthResponse(r *ghttp.Request, status string, checks map[string]bool) {
	data := g.Map{
		"status": status,
		"checks": checks,
		"timestamp": g.NewVar(nil).Time(),
	}
	SuccessResponse(r, data)
}