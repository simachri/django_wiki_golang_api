package models

import "fmt"

// Resource is the result of an API call.
type Resource interface {
	// Equals returns 'True' if the contents of the provided Response equals this Response instance's contents.
	Equals(r Resource) bool
	// GetPath returns the path to the resource.
	GetPath() string
	// IsRoot returns 'True' if it is the root article.
	IsRoot() bool
}

// ArticleBase is the common base of every article.
type ArticleBase struct {
	ID         int    `json:"id"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	RevisionID int    `json:"revision_id" db:"rev_id"`
	// ParentArtID is the value of wiki_article-id of the parent node.
	// Note: The database model uses wiki_urlpath-parent_id = wiki_urlpath-id for
	// the hierarchy, see ParentPathID below..
	// ParentArtID has no 'required' binding as the handler of POST /articles does not know
	// whether a root or a child article is to be created/updated.
	ParentArtID int `json:"parent_art_id" db:"parent_art_id"`
	// PathID is the value of wiki_urlpath-id.
	PathID int `json:"path_id" db:"path_id"`
    Left   int `json:"left" db:"lft"`
	Right  int `json:"right" db:"rght"`
}

// RootArticle is the root Wiki article.
type RootArticle struct {
	ArticleBase
}

// Article is a non-root Wiki article.
type Article struct {
	ArticleBase
	Slug  string `json:"slug"`
    Level int    `json:"level" db:"level"`
}

// Equals returns 'True' if the contents of the provided RootArticle equals this RootArticle instance's contents.
func (a RootArticle) Equals(r Resource) bool {
	b, ok := r.(*RootArticle)
	if !ok {
		return false
		// Do not compare IDs for the sake of easier testing.
		// TODO: Find a better approach.
		//} else if a.ID != b.ID {
		//return false
	} else if a.Title != b.Title {
		return false
	} else if a.Content != b.Content {
		return false
	}
	return true
}

// Equals returns 'True' if the contents of the provided RootArticle equals this RootArticle instance's contents.
func (a Article) Equals(r Resource) bool {
	b, ok := r.(*Article)
	if !ok {
		return false
		// Do not compare IDs for the sake of easier testing.
		// TODO: Find a better approach.
		//} else if a.ID != b.ID {
		//return false
	} else if a.Title != b.Title {
		return false
	} else if a.Content != b.Content {
		return false
	} else if a.Slug != b.Slug {
		return false
		// Do not compare IDs for the sake of easier testing.
		// TODO: Find a better approach.
		//} else if a.RevisionID != b.RevisionID {
		//return false
		//} else if a.ParentID != b.ParentID {
		//return false
		//}
	}
	return true
}

// GetPath returns the path to the resource.
func (a ArticleBase) GetPath() string {
	return fmt.Sprintf("articles/%v", a.ID)
}

// IsRoot returns 'True' if it is the root article.
func (a Article) IsRoot() bool {
	return false
}

// IsRoot returns 'True' if it is the root article.
func (a RootArticle) IsRoot() bool {
	return true
}
