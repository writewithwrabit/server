package server

type Editor struct {
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	ShowToolbar bool   `json:"showToolbar"`
	ShowPrompt  bool   `json:"showPrompt"`
	ShowCounter bool   `json:"showCounter"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}
