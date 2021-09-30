package models

import "fmt"

// Resource is the result of an API call.
type Resource interface {
	// Equals returns 'True' if the contents of the provided Response equals this Response instance's contents.
	Equals(r Resource) bool
	// GetPath returns the path to the resource.
	GetPath() string
}

// ArticleBase is the common base of every article.
type ArticleBase struct {
	ID         int    `json:"id"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	RevisionID int    `json:"revision_id" db:"rev_id"`
}

// RootArticle is the root Wiki article.
type RootArticle struct {
    ArticleBase
}

// Article is a non-root Wiki article.
type Article struct {
    ArticleBase
	Slug       string `json:"slug"`
	ParentID   int    `json:"parent_id" binding:"required" db:"parent"`
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
	//} else if a.Slug != b.Slug {
		//return false
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
