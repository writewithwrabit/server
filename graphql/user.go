package server

type User struct {
	ID                   string  `json:"id"`
	FirebaseID           *string `json:"firebaseID"`
	StripeID             *string `json:"stripeID"`
	FirstName            string  `json:"firstName"`
	LastName             *string `json:"lastName"`
	Email                string  `json:"email"`
	WordGoal             int     `json:"wordGoal"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`
	StripeSubscriptionID string  `json:"stripeSubscriptionID"`
}
