package models

type Streak struct {
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	DayCount    int    `json:"dayCount"`
	LastEntryID string `json:"lastEntryId"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}
