package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	models "coco-life.de/wapi/internal"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

func dbHealthCheck(c *gin.Context) {
	/* The database connection parameters will be loaded from environment variables.
	 * user=<PGUSER> host=<PGHOST> password=<PGPASSWORD> port=<PGPORT>
	 * dbname=<PGDATABASE>
	 * The mapping of environment variables to keyboard is as follows:
	 * hostaddr -> PGHOST
	 * port -> PGPORT
	 * user -> PGUSER
	 * password -> PGPASSWORD
	 * dbname -> PGDATABASE
	 * See `go doc pgconn.ParseConfig` for details.
	 */
	dbpool, err := pgxpool.Connect(context.Background(), "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	var greeting string
	err = dbpool.QueryRow(context.Background(), "select 'Hello, world!';").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	c.String(200, fmt.Sprintln(greeting)+"Database connection up and running.")
}

func fetchArticleByID(c *gin.Context) {
	dbpool, err := pgxpool.Connect(context.Background(), "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	var article models.Article
	err = pgxscan.Get(
		context.Background(), dbpool, &article,
		`select
            hdr.id,
            rev.title,
            rev.content
        from wiki_article as hdr
            inner join wiki_articlerevision as rev
                on hdr.id = rev.article_id
        where hdr.id = $1;`, c.Param("id"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to query database table wiki_article: %v\n", err)
		os.Exit(1)
	}

	c.JSON(http.StatusOK, article)
}

// https://github.com/gin-gonic/gin#testing
func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	r.GET("/db/health", dbHealthCheck)
	r.GET("/articles/:id", fetchArticleByID)
	return r
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
