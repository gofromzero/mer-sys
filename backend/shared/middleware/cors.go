package middleware

import (
	"github.com/gogf/gf/v2/net/ghttp"
)

// CORS handles cross-origin resource sharing
func CORS() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		r.Response.CORSDefault()
		r.Middleware.Next()
	}
}