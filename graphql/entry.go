package server

type Entry struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	WordCount int    `json:"wordCount"`
	Content   string `json:"content"`
	GoalHit   bool   `json:"goalHit"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
