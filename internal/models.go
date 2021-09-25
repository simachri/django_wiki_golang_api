package models

import "fmt"

// Resource is the result of an API call.
type Resource interface {
	// Equals returns 'True' if the contents of the provided Response equals this Response instance's contents.
	Equals(r Resource) bool
	// GetPath returns the path to the resource.
	GetPath() string
}

// Article is a Wiki article.
type Article struct {
	ID         int    `json:"id"`
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content"`
	Slug       string `json:"slug" binding:"required"`
	RevisionID int    `json:"revision_id"`
	ParentID   int    `json:"parent_id" binding:"required"`
}

// Equals returns 'True' if the contents of the provided Response equals this Response instance's contents.
func (a Article) Equals(r Resource) bool {
	b, ok := r.(*Article)
	if !ok {
		return false
	} else if a.ID != b.ID {
		return false
	} else if a.Title != b.Title {
		return false
	} else if a.Content != b.Content {
		return false
	}
	return true
}

// GetPath returns the path to the resource.
func (a Article) GetPath() string {
	return fmt.Sprintf("articles/%v", a.ID)
}
