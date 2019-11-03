// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package server

type ExistingEntry struct {
	UserID    string `json:"userID"`
	WordCount int    `json:"wordCount"`
	Content   string `json:"content"`
	GoalHit   bool   `json:"goalHit"`
}

type NewEditor struct {
	UserID      string `json:"userId"`
	ShowToolbar bool   `json:"showToolbar"`
	ShowPrompt  bool   `json:"showPrompt"`
	ShowCounter bool   `json:"showCounter"`
}

type NewEntry struct {
	UserID    string `json:"userId"`
	WordCount int    `json:"wordCount"`
	Content   string `json:"content"`
}

type NewSubscription struct {
	StripeID       string `json:"stripeId"`
	TokenID        string `json:"tokenId"`
	SubscriptionID string `json:"subscriptionId"`
}

type NewUser struct {
	FirstName string  `json:"firstName"`
	LastName  *string `json:"lastName"`
	Email     string  `json:"email"`
}

type PreferredWritingTime struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

type Stats struct {
	WordsWritten          int                     `json:"wordsWritten"`
	LongestStreak         int                     `json:"longestStreak"`
	LongestEntry          int                     `json:"longestEntry"`
	PreferredWritingTimes []*PreferredWritingTime `json:"preferredWritingTimes"`
	PreferredDayOfWeek    int                     `json:"preferredDayOfWeek"`
}

type UpdatedUser struct {
	ID         string  `json:"id"`
	FirebaseID *string `json:"firebaseID"`
	StripeID   *string `json:"stripeID"`
	FirstName  *string `json:"firstName"`
	LastName   *string `json:"lastName"`
	Email      *string `json:"email"`
	WordGoal   *int    `json:"wordGoal"`
}

type User struct {
	ID         string  `json:"id"`
	FirebaseID *string `json:"firebaseID"`
	StripeID   *string `json:"stripeID"`
	FirstName  string  `json:"firstName"`
	LastName   *string `json:"lastName"`
	Email      string  `json:"email"`
	WordGoal   int     `json:"wordGoal"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
}
