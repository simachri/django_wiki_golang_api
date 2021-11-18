package db

import (
	"context"
	"fmt"

	"coco-life.de/wapi/internal/models"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
)

// SelectRootArticle selects the root article from the database.
func SelectRootArticle(dbpool *pgxpool.Pool) (*models.RootArticle, error) {
	var article models.RootArticle
	err := pgxscan.Get(
		context.Background(), dbpool, &article,
		`select
            hdr.id,
            rev.id as rev_id,
            rev.title,
            rev.content
        from wiki_article as hdr
            inner join wiki_articlerevision as rev
                on hdr.id = rev.article_id
            inner join wiki_urlpath as path
                on hdr.id = path.article_id  
        where path.level = 0;`)
	return &article, err
}

// SelectArticleBySlug selects a specific article by its slug.
func SelectArticleBySlug(dbpool *pgxpool.Pool, slug string) (*models.Article, error) {
	var article models.Article
	err := pgxscan.Get(
		context.Background(), dbpool, &article,
		`select
            hdr.id,
            rev.id as rev_id,
            rev.title,
            rev.content,
            path.slug,
            COALESCE(path.parent_id, -1) as parent
        from wiki_article as hdr
            inner join wiki_articlerevision as rev
                on hdr.id = rev.article_id
            inner join wiki_urlpath as path
                on hdr.id = path.article_id  
        where path.slug = $1;`, slug)
	return &article, err
}

// InsertWikiURLPathRoot inserts the record into wiki_urlpath for the root article.
func InsertWikiURLPathRoot(conn *pgxpool.Pool, hdrID int) error {
    // TODO: Adjust lft and rght.
	sql := `insert into
      wiki_urlpath
      (
        lft,
        rght,
        level,
        tree_id,
        article_id,
        site_id
      )
      values
      (
        1,
        2,
        0,
        1,
        $1,
        1
      )`
	var commandTag pgconn.CommandTag
	var err error
    commandTag, err = conn.Exec(context.Background(), sql, hdrID)
	if err != nil {
		return fmt.Errorf("Failed to insert record into wiki_urlpath: %v", err)
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("Failed to insert record into wiki_urlpath")
	}
	return nil
}

// InsertWikiURLPath inserts the record into wiki_urlpath for a non-root article.
func InsertWikiURLPath(conn *pgxpool.Pool, hdrID int, slug string, parentID int) error {
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
    commandTag, err = conn.Exec(context.Background(), sql, hdrID, slug, parentID)
	if err != nil {
		return fmt.Errorf("Failed to insert record into wiki_urlpath: %v", err)
	}
	if commandTag.RowsAffected() != 1 {
		return fmt.Errorf("Failed to insert record into wiki_urlpath")
	}
	return nil
}

// InsertWikiArticleRevision creates the record in wiki_articlerevision.
func InsertWikiArticleRevision(conn *pgxpool.Pool, hdrID int, title string, content string) (int, error) {
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

// InsertWikiArticle a record into wiki_article.
func InsertWikiArticle(conn *pgxpool.Pool) (int, error) {
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

// SetWikiArticleRevision database table wiki_article and sets the revision.
func SetWikiArticleRevision(conn *pgxpool.Pool, hdrID int, revID int) error {
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


