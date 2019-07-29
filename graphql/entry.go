package server

type Entry struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	WordCount int    `json:"wordCount"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
