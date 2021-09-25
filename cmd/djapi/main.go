package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	models "coco-life.de/wapi/internal"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
)

var baseURL string

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

	c.String(http.StatusOK, fmt.Sprintln(greeting)+"Database connection up and running.")
}

func fetchArticle(c *gin.Context) {
	dbpool, err := pgxpool.Connect(context.Background(), "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	id, err := strconv.Atoi(c.Param("id"))
	if notOK := handleErr(c, &err, "Failed to parse GET parameter 'id' as integer: %v\n"); notOK {
		return
	}

	article, err := selectArticleByID(dbpool, id)
	if notOK := handleErr(c, &err, "Failed to query database table wiki_article: %v\n"); notOK {
		return
	}

	c.JSON(http.StatusOK, article)
}

func selectArticleByID(dbpool *pgxpool.Pool, id int) (*models.Article, error) {
	var article models.Article
	err := pgxscan.Get(
		context.Background(), dbpool, &article,
		`select
            hdr.id,
            rev.title,
            rev.content
        from wiki_article as hdr
            inner join wiki_articlerevision as rev
                on hdr.id = rev.article_id
        where hdr.id = $1;`, id)
	return &article, err
}

func insertWikiURLPath(conn *pgxpool.Pool, hdrID int, slug string, parentID int) error {
	sql := `insert into
      wiki_urlpath
      (
        slug,
        lft,
        rght,
        level,
        tree_id,
        article_id,
        site_id,
        parent_id
      )
      values
      (
        $2,
        2,
        3,
        1,
        1,
        $1,
        1,
        $3
      )`
	var commandTag pgconn.CommandTag
	var err error
	if parentID == -1 {
		commandTag, err = conn.Exec(context.Background(), sql, hdrID, slug, nil)
	} else {
		commandTag, err = conn.Exec(context.Background(), sql, hdrID, slug, parentID)
    }
	if err != nil {
		return fmt.Errorf("Failed to insert record into wiki_urlpath: %v", err)
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("Failed to insert record into wiki_urlpath")
	}
	return nil
}
func insertWikiArticleRevision(conn *pgxpool.Pool, hdrID int, title string, content string) (int, error) {
	sql := `insert into
      wiki_articlerevision
      (
        article_id,
        revision_number,
        previous_revision_id,
        title,
        content,
        created,
        modified,
        deleted,
        locked,
        user_message,
        automatic_log
      )
      values 
      (
        $1,
        1,
        null,
        $2,
        $3,
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        false,
        false,
        '',
        ''
      )
      returning id as rev_id;`
	row := conn.QueryRow(context.Background(), sql, hdrID, title, content)
	var revID int
	err := row.Scan(&revID)
	if err != nil {
		return -1, fmt.Errorf("Failed to insert record into wiki_articlerevision: %v", err)
	}
	return revID, nil
}

func insertWikiArticle(conn *pgxpool.Pool) (int, error) {
	sql := `insert into
      wiki_article
      (
        created,
        modified,
        group_read,
        group_write,
        other_read,
        other_write,
        current_revision_id
      )
      values
      (
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP,
        true,
        true,
        true,
        true,
        null -- revision_id has a UNIQUE constraint. We can set it once the revision is created.
      )
      returning id as hdr_id;`
	row := conn.QueryRow(context.Background(), sql)
	var hdrID int
	err := row.Scan(&hdrID)
	if err != nil {
		return -1, fmt.Errorf("Failed to insert record into wiki_articlerevision: %v", err)
	}
	return hdrID, nil
}

func setWikiArticleRevision(conn *pgxpool.Pool, hdrID int, revID int) error {
	sql := `update wiki_article
                set current_revision_id = $2
                where id = $1;`
	commandTag, err := conn.Exec(context.Background(), sql, hdrID, revID)
	if err != nil {
		return fmt.Errorf("Failed to update 'current_revision_id' in wiki_article: %v", err)
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("Failed to update 'current_revision_id' in wiki_article")
	}
	return nil
}

func addArticle(c *gin.Context) {
	var articleInput models.Article
	if err := c.ShouldBindJSON(&articleInput); err != nil {
		if notOK := handleErr(c, &err, "addArticle: Failed to bind 'Article': %v\n"); notOK {
			return
		}
	}

	dbpool, err := pgxpool.Connect(context.Background(), "")
	if notOK := handleErr(c, &err, "addArticle: Unable to connect to database: %v\n"); notOK {
		return
	}
	defer dbpool.Close()

	tx, err := dbpool.Begin(context.Background())
	if notOK := handleErr(c, &err, "addArticle: Failed to create transaction: %v\n"); notOK {
		return
	}

	hdrID, err := insertWikiArticle(dbpool)
	if notOK := handleErr(c, &err, "addArticle: Failed to INSERT into wiki_article: %v\n"); notOK {
		return
	}

	revID, err := insertWikiArticleRevision(dbpool, hdrID, articleInput.Content, articleInput.Title)
	if notOK := handleErr(c, &err, "addArticle: Failed to INSERT into wiki_articlerevision: %v\n"); notOK {
		return
	}

	err = insertWikiURLPath(dbpool, hdrID, articleInput.Slug, articleInput.ParentID)
	if notOK := handleErr(c, &err, "addArticle: Failed to INSERT into wiki_articlerevision: %v\n"); notOK {
		return
	}

	setWikiArticleRevision(dbpool, hdrID, revID)

	err = tx.Commit(context.Background())
	if notOK := handleErr(c, &err, "addArticle: Failed to commit transaction to insert new article: %v\n"); notOK {
		return
	}

	articleOut, err := selectArticleByID(dbpool, hdrID)
	if notOK := handleErr(c, &err, "Failed to query database table wiki_article: %v\n"); notOK {
		return
	}
	c.JSON(http.StatusCreated, articleOut)
	c.Header("Location", buildResourceURL(baseURL, articleOut))
}

// buildResourceURL creates a fully qualified URL to access the resource.
func buildResourceURL(baseURL string, r models.Resource) string {
	return baseURL + r.GetPath()
}

// handleErr returns 'true' if an error has been handled.
func handleErr(c *gin.Context, e *error, m string) bool {
	if *e == nil {
		return false
	}
	msg := fmt.Sprintf(m, *e)
	fmt.Fprint(os.Stderr, msg)
	c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	return true
}

// https://github.com/gin-gonic/gin#testing
func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.GET("/db/health", dbHealthCheck)
	r.GET("/articles/:id", fetchArticle)
	r.POST("/articles", addArticle)
	return r
}

func readEnv() {
	// Load the .env file in the current directory
	godotenv.Load()
	baseURL = "https://" + os.Getenv("host") + "/"

}

func main() {
	r := setupRouter()
	r.Run(":8080")
	readEnv()
}
