package models

import stripe "github.com/stripe/stripe-go"

type StripeSubscription struct {
	ID               string                    `json:"id"`
	CurrentPeriodEnd int64                     `json:"currentPeriodEnd"`
	TrialEnd         int64                     `json:"trialEnd"`
	CancelAt         int64                     `json:"cancelAt"`
	Status           stripe.SubscriptionStatus `json:"status"`
	Plan             *stripe.Plan              `json:"plan"`
}
