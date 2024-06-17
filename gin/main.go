package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	server := gin.Default()
	server.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello go")
	})

	server.Run(":8080") // 监听并在 0.0.0.0:8080 上启动服务
}
