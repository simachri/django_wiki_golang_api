package models

// Response is the result of an API call.
type Response interface {
	// Equals returns 'True' if the contents of the provided Response equals this Response instance's contents.
	Equals(r Response) bool
}

// Article is a Wiki article.
type Article struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Equals returns 'True' if the contents of the provided Response equals this Response instance's contents.
func (a Article) Equals(r Response) bool {
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
