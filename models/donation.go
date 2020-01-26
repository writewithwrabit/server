package models

type Donation struct {
	ID          string `json:"id"`
	UserID      string `json:"userId"`
	Amount      int    `json:"amount"`
	Paid        bool   `json:"paid"`
	EntryID     string `json:"entryId"`
	LastEntryID string `json:"lastEntryId"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}
