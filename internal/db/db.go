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
            rev.content,
            path.lft,
            path.rght
        from wiki_article as hdr
            inner join wiki_articlerevision as rev
                on hdr.id = rev.article_id
            inner join wiki_urlpath as path
                on hdr.id = path.article_id  
        where path.level = 0;`)
	return &article, err
}

// SelectArticleByID selects a specific article by wiki_article-id.
func SelectArticleByID(dbpool *pgxpool.Pool, id int) (*models.Article, error) {
	var article models.Article
	err := pgxscan.Get(
		context.Background(), dbpool, &article,
		`select
            hdr.id,
            rev.id as rev_id,
            rev.title,
            rev.content,
            COALESCE(path.slug, '') as slug,
            path.id as path_id,
            path.level,
            path.lft,
            path.rght,
            COALESCE(parent_hdr.id, -1) as parent_art_id
        from wiki_article as hdr
            inner join wiki_articlerevision as rev
                on hdr.id = rev.article_id
            inner join wiki_urlpath as path
                on hdr.id = path.article_id  
            left join wiki_urlpath as parent_path
                on path.parent_id = parent_path.id  
            left join wiki_article as parent_hdr
                on parent_path.article_id = parent_hdr.id
        where hdr.id = $1;`, id)
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
            COALESCE(path.parent_id, -1) as parent_path_id
        from wiki_article as hdr
            inner join wiki_articlerevision as rev
                on hdr.id = rev.article_id
            inner join wiki_urlpath as path
                on hdr.id = path.article_id  
        where path.slug = $1;`, slug)
	return &article, err
}

// MPTTCalcForIns calculates the 'level', 'left' and 'right' for a node under a parent.
// prtRgh is the 'right' value of the parent. The 'right' value is the anchor for being 
// able to add new child nodes as right siblings to other already existing children.
func MPTTCalcForIns(prtLvl int, prtRght int) (lvl int, left int, right int) {
    // Insert a new article `n` as child to parent `p`:
    // Set `lft` and `rght` of `n` based on `p`.
    // - `n.lft = p.rght`
    // - `n.rght = p.rght + 1`
    chLft := prtRght
    chRght := prtRght + 1
    chLvl := prtLvl + 1
    return chLvl, chLft, chRght
}

// MPTTUpdWikiURLPathForInsert updates all wiki_urlpath records after another node has been 
// inserted.
// Adjust `lft` and `rght` of all nodes `r` that are
// - either right siblings to `n` (including their children)
// - or direct children of `n`
// - or direct parent
// - or ancestors (parent and grandparent of parent)
// - or right to direct parent or ancestors.
//
// All their `lft` and `rght` values need to be incremented by `2`:
// - `lft`: All nodes `r` with `r.lft >= n.lft`:
//   `r.lft = r.lft + 2`
// - `rght`: All nodes `r` with `r.rght >= n.lft`:
//   `r.rght = r.rght + 2`
//
// Note: The condition `r.rght >= r.rght` (mind the `rght` instead of the 
// `lft`) does not cover for parent nodes as their `r.rght` is not matched by 
// this condition. Example: Parent node has `r.lft = 1 and r.rght = 2`. New 
// node is inserted with `n.lft = 2 and n.rght = 3`. `r.rght` has to be set to 
// `4`.
func MPTTUpdWikiURLPathForInsert(conn *pgxpool.Pool, newArtPathID, nLft int) error {
	var err error
	sqlUpdLft := `update wiki_urlpath
        set lft = lft + 2
        where lft >= $1
              and not id = $2
              `
	_, err = conn.Exec(context.Background(), sqlUpdLft, nLft, newArtPathID)
	if err != nil {
		return fmt.Errorf("Failed to update record in wiki_urlpath: %v", err)
	}

    // These two SQL statements cannot be merged into one as for some nodes, e.g. 
    // parents, only the field 'rght' needs to be updated.
	sqlUpdRght := `update wiki_urlpath
        set rght = rght + 2
        where rght >= $1
              and not id = $2
               `
	_, err = conn.Exec(context.Background(), sqlUpdRght, nLft, newArtPathID)
	if err != nil {
		return fmt.Errorf("Failed to update record in wiki_urlpath: %v", err)
	}

	return nil
}

// InsertWikiURLPathChild inserts the record into wiki_urlpath for any child article.
// parentPathId is the value of wiki_urlpath-id of the parent's node.
// It returns wiki_urlpath-id.
func InsertWikiURLPathChild(conn *pgxpool.Pool,
                            slug string,
                            hdrID int,
                            lvl int,
                            left int,
                            right int,
                            parentPathID int,
                            ) (int, error) {
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
        $1,
        $2,
        $3,
        $4,
        1,
        $5,
        1,
        $6
      )
      returning id`
      row := conn.QueryRow(context.Background(),
                                sql,
                                slug,
                                left,
                                right,
                                lvl,
                                hdrID,
                                parentPathID)
	var pathID int
	err := row.Scan(&pathID)
	if err != nil {
		return -1, fmt.Errorf("Failed to insert record into wiki_urlpath: %v", err)
	}
	return pathID, nil
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
// It returns wiki_articlerevision-id.
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
