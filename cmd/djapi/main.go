package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
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
	conn, err := pgx.Connect(context.Background(), "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	var greeting string
	err = conn.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	c.String(200, fmt.Sprintln(greeting)+"Database connection up and running.")
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
