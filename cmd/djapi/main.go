package main

import (
	"github.com/gin-gonic/gin"
)

func dbHealthCheck(c *gin.Context) {

}

// https://github.com/gin-gonic/gin#testing
func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	r.GET("/db/health", dbHealthCheck)
	return r
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
