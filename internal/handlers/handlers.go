package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"coco-life.de/wapi/internal/db"
	"coco-life.de/wapi/internal/models"
	"coco-life.de/wapi/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
)

var baseURL string

// FetchRootArticle selects the root article from the database.
func FetchRootArticle(c *gin.Context) {
	dbpool, err := pgxpool.Connect(context.Background(), "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	article, err := db.SelectRootArticle(dbpool)
	if notOK := utils.HandleErr(c, &err, "Failed to query database table wiki_article: %v\n"); notOK {
		return
	}

	c.JSON(http.StatusOK, article)
}

// FetchArticleBySlug returns an article given by its slug.
func FetchArticleBySlug(c *gin.Context) {
	dbpool, err := pgxpool.Connect(context.Background(), "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	article, err := db.SelectArticleBySlug(dbpool, c.Param("slug"))
	if notOK := utils.HandleErr(c, &err, "Failed to query database table wiki_article: %v\n"); notOK {
		return
	}

	c.JSON(http.StatusOK, article)
}

// AddRootArticle adds/sets the root article.
func AddRootArticle(c *gin.Context) {
	var articleInput models.RootArticle
	if err := c.ShouldBindJSON(&articleInput); err != nil {
		if notOK := utils.HandleErr(c, &err, "addArticle: Failed to bind 'Article': %v\n"); notOK {
			return
		}
	}

	dbpool, err := pgxpool.Connect(context.Background(), "")
	if notOK := utils.HandleErr(c, &err, "addArticle: Unable to connect to database: %v\n"); notOK {
		return
	}
	defer dbpool.Close()

	tx, err := dbpool.Begin(context.Background())
	if notOK := utils.HandleErr(c, &err, "addArticle: Failed to create transaction: %v\n"); notOK {
		return
	}

	hdrID, err := db.InsertWikiArticle(dbpool)
	if notOK := utils.HandleErr(c, &err, "addArticle: Failed to INSERT into wiki_article: %v\n"); notOK {
		return
	}

	revID, err := db.InsertWikiArticleRevision(dbpool, hdrID, articleInput.Title, articleInput.Content)
	if notOK := utils.HandleErr(c, &err, "addArticle: Failed to INSERT into wiki_articlerevision: %v\n"); notOK {
		return
	}

	err = db.InsertWikiURLPathRoot(dbpool, hdrID)
	if notOK := utils.HandleErr(c, &err, "addArticle: Failed to INSERT into wiki_urlpath: %v\n"); notOK {
		return
	}

	db.SetWikiArticleRevision(dbpool, hdrID, revID)

	err = tx.Commit(context.Background())
	if notOK := utils.HandleErr(c, &err, "addArticle: Failed to commit transaction to insert new article: %v\n"); notOK {
		return
	}

	articleOut, err := db.SelectRootArticle(dbpool)
	if notOK := utils.HandleErr(c, &err, "Failed to query database table wiki_article: %v\n"); notOK {
		return
	}
	c.JSON(http.StatusCreated, articleOut)
	c.Header("Location", buildResourceURL(baseURL, articleOut))
}

// DbHealthCheck returns HTTP 200 if the database connection works.
func DbHealthCheck(c *gin.Context) {
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

	c.String(http.StatusOK, fmt.Sprintln(greeting)+"Database connection up and running.")
}

// buildResourceURL creates a fully qualified URL to access the resource.
func buildResourceURL(baseURL string, r models.Resource) string {
	return baseURL + r.GetPath()
}

// SetBaseURL sets the valur of baseURL, that is, the domain of the API.
func SetBaseURL(new string) {
    baseURL = new
}
