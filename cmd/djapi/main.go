package main

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"coco-life.de/wapi/internal/handlers"
	"github.com/gin-gonic/gin"
)

// https://github.com/gin-gonic/gin#testing
func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.GET("/db/health", handlers.DbHealthCheck)
	r.POST("/articles", handlers.InsertArticle)
	r.GET("/articles", handlers.FetchRootArticle)
    r.GET("/articles/:id", handlers.RetrieveArticleByID)
	return r
}

func readEnv() {
	// Load the .env file in the current directory
	godotenv.Load()
	handlers.SetBaseURL("https://" + os.Getenv("host") + "/")
}

func main() {
	r := setupRouter()
	r.Run(":8080")
	readEnv()
}
