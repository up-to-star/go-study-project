package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}
		//if ctx.Request.URL.Path == "/users/login" ||
		//	ctx.Request.URL.Path == "/users/signup" {
		//	return
		//}
		sess := sessions.Default(ctx)
		id := sess.Get("userId")
		sess.Set("userId", id)
		sess.Options(sessions.Options{
			MaxAge: 60,
		})
		if id == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		updateTime := sess.Get("update_time")
		now := time.Now().UnixMilli()
		if updateTime == nil {
			sess.Set("update_time", now)
			sess.Save()
			return
		}
		updateTimeval, _ := updateTime.(int64)
		if now-updateTimeval >= 60*60*1000 {
			sess.Set("update_time", now)
			sess.Save()
		}
	}
}

func (l *LoginMiddlewareBuilder) IgnorePaths(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}
