//go:generate go run github.com/99designs/gqlgen

package resolvers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/mailgun/mailgun-go/v3"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/card"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/sub"
	"github.com/writewithwrabit/server/auth"
	wrabitDB "github.com/writewithwrabit/server/db"
	"github.com/writewithwrabit/server/graph/generated"
	"github.com/writewithwrabit/server/models"
)

type Resolver struct {
	db      *sql.DB
	editors []*models.Editor
	entries []*models.Entry
}

func New(db *sql.DB) generated.Config {
	return generated.Config{
		Resolvers: &Resolver{
			db: db,
		},
	}
}

func (r *Resolver) Mutation() generated.MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{r}
}

func (r *Resolver) Editor() generated.EditorResolver {
	return &editorResolver{r}
}

func (r *Resolver) Entry() generated.EntryResolver {
	return &entryResolver{r}
}

func (r *Resolver) Streak() generated.StreakResolver {
	return &streakResolver{r}
}

func (r *Resolver) User() generated.UserResolver {
	return &userResolver{r}
}

func (r *Resolver) StripeSubscription() generated.StripeSubscriptionResolver {
	return &stripeSubscriptionResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateUser(ctx context.Context, input models.NewUser) (*models.User, error) {
	// Initialize Stripe
	stripe.Key = os.Getenv("STRIPE_KEY")

	user := &models.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
	}

	params := &stripe.CustomerParams{
		Name:  stripe.String(user.FirstName),
		Email: stripe.String(user.Email),
	}
	cus, err := customer.New(params)
	if err != nil {
		panic(err)
	}

	// Add the Stripe ID so that it returns
	user.StripeID = &cus.ID

	res := wrabitDB.LogAndQueryRow(r.db, "INSERT INTO users (first_name, last_name, email, stripe_id) VALUES ($1, $2, $3, $4) RETURNING id", user.FirstName, user.LastName, user.Email, cus.ID)
	if err := res.Scan(&user.ID); err != nil {
		panic(err)
	}

	return user, nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, input models.UpdatedUser) (*models.User, error) {
	res := wrabitDB.LogAndQueryRow(r.db, "SELECT id, firebase_id, stripe_id, first_name, last_name, email, word_goal FROM users WHERE id = $1", input.ID)

	// TODO: Figure out why createdAt and updatedAt didn't work on this query
	var user models.User
	if err := res.Scan(&user.ID, &user.FirebaseID, &user.StripeID, &user.FirstName, &user.LastName, &user.Email, &user.WordGoal); err != nil {
		panic(err)
	}

	userContext := auth.ForContext(ctx)
	if userContext == nil || userContext.Subject != *user.FirebaseID {
		return &models.User{}, fmt.Errorf("Access denied")
	}

	firebaseID := user.FirebaseID
	if input.FirebaseID != nil {
		firebaseID = input.FirebaseID
	}

	stripeID := user.StripeID
	if input.StripeID != nil {
		stripeID = input.StripeID
	}

	firstName := user.FirstName
	if input.FirstName != nil {
		firstName = *input.FirstName
	}

	lastName := user.LastName
	if input.LastName != nil {
		lastName = input.LastName
	}

	email := user.Email
	if input.Email != nil {
		email = *input.Email
	}

	wordGoal := user.WordGoal
	if input.WordGoal != nil {
		wordGoal = *input.WordGoal
	}

	user = models.User{
		ID:         input.ID,
		FirebaseID: firebaseID,
		StripeID:   stripeID,
		FirstName:  firstName,
		LastName:   lastName,
		Email:      email,
		WordGoal:   wordGoal,
	}

	res = wrabitDB.LogAndQueryRow(r.db, "UPDATE users SET firebase_id = $1, stripe_id = $2, first_name = $3, last_name = $4, email = $5, word_goal = $6 WHERE id = $7 RETURNING id", user.FirebaseID, user.StripeID, user.FirstName, user.LastName, user.Email, user.WordGoal, user.ID)
	if err := res.Scan(&user.ID); err != nil {
		panic(err)
	}

	return &user, nil
}

func (r *mutationResolver) CompleteUserSignup(ctx context.Context, input models.SignedUpUser) (*models.User, error) {
	res := wrabitDB.LogAndQueryRow(r.db, "SELECT id, firebase_id, stripe_id, first_name, last_name, email, word_goal FROM users WHERE id = $1", input.ID)
	var user models.User
	if err := res.Scan(&user.ID, &user.FirebaseID, &user.StripeID, &user.FirstName, &user.LastName, &user.Email, &user.WordGoal); err != nil {
		panic(err)
	}

	res = wrabitDB.LogAndQueryRow(r.db, "UPDATE users SET firebase_id = $1 WHERE id = $2 RETURNING id", input.FirebaseID, input.ID)
	if err := res.Scan(&user.ID); err != nil {
		panic(err)
	}

	// Initialize Mailgun
	mgKey := os.Getenv("MAILGUN_KEY")
	mg := mailgun.NewMailgun("mg.writewithwrabit.com", mgKey)

	sender := "Team Wrabit <hello@writewithwrabit.com>"
	subject := "Welcome to your writing journey!"
	body := ""
	recipient := user.Email

	message := mg.NewMessage(sender, subject, body, recipient)
	message.SetTemplate("app-template")
	message.AddTemplateVariable("content", `Hey there! 👋<br><br>
  
  We hope you're ready to build a daily writing habit. It might not be easy but it's definitely rewarding!
  We have a few tips to help you get started.<br><br>

  1. <b>Don't think too much.</b> Let whatever needs to come out, come out.<br>
  2. <b>Don't feel to bad if you miss a day.</b> At Wrabit we start small and every word counts.<br>
  3. <b>Have fun! 🎉</b> Building a habit is hard so we want it to be as enjoyable as possible.<br><br>

  If there is anything we can do to support you, feel free to reach out. You can respond directly to this email! Our platform is new but we have lots planned. Thanks for being apart of <em>our</em> journey.<br><br>

  Be well,<br>
  Team Wrabit 🐇
  `)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err := mg.Send(ctx, message)

	if err != nil {
		panic(err)
	}

	return &user, nil
}

func (r *mutationResolver) CreateSubscription(ctx context.Context, input models.NewSubscription) (*models.StripeSubscription, error) {
	// Initialize Stripe
	stripe.Key = os.Getenv("STRIPE_KEY")

	cardParams := &stripe.CardParams{
		Customer: stripe.String(input.StripeID),
		Token:    stripe.String(input.TokenID),
	}

	_, err := card.New(cardParams)
	if err != nil {
		panic(err)
	}

	subParams := &stripe.SubscriptionParams{
		Customer: stripe.String(input.StripeID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: stripe.String(input.SubscriptionID),
			},
		},
		TrialFromPlan: stripe.Bool(input.Trial),
	}

	subscription, err := sub.New(subParams)
	if err != nil {
		panic(err)
	}

	var user = models.User{
		StripeID: &input.StripeID,
	}
	res := wrabitDB.LogAndQueryRow(r.db, "UPDATE users SET stripe_subscription_id = $1 WHERE stripe_id = $2 RETURNING id", subscription.ID, input.StripeID)
	if err := res.Scan(&user.ID); err != nil {
		panic(err)
	}

	var newSubscription = &models.StripeSubscription{
		ID:               subscription.ID,
		CurrentPeriodEnd: subscription.CurrentPeriodEnd,
		TrialEnd:         subscription.TrialEnd,
		CancelAt:         subscription.CancelAt,
		Status:           subscription.Status,
		Plan:             subscription.Plan,
	}

	return newSubscription, nil
}

func (r *mutationResolver) CancelSubscription(ctx context.Context, id string) (string, error) {
	// Initialize Stripe
	stripe.Key = os.Getenv("STRIPE_KEY")

	_, err := sub.Cancel(id, nil)
	if err != nil {
		panic(err)
	}

	return "ok", nil
}

func (r *mutationResolver) CreateEditor(ctx context.Context, input models.NewEditor) (*models.Editor, error) {
	if user := auth.ForContext(ctx); user == nil {
		return &models.Editor{}, fmt.Errorf("Access denied")
	}

	editor := &models.Editor{
		ID:          "",
		UserID:      input.UserID,
		ShowToolbar: input.ShowToolbar,
		ShowPrompt:  input.ShowPrompt,
		ShowCounter: input.ShowCounter,
	}

	res := wrabitDB.LogAndQueryRow(r.db, "INSERT INTO editors (user_id, show_toolbar, show_prompt, show_counter) VALUES ($1, $2, $3, $4) RETURNING id", editor.UserID, editor.ShowToolbar, editor.ShowPrompt, editor.ShowCounter)
	if err := res.Scan(&editor.ID); err != nil {
		panic(err)
	}

	return editor, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) User(ctx context.Context, id *string) (*models.User, error) {
	if user := auth.ForContext(ctx); user == nil {
		return &models.User{}, fmt.Errorf("Access denied")
	}

	res := wrabitDB.LogAndQueryRow(r.db, "SELECT id, firebase_id, stripe_id, first_name, last_name, email, word_goal, stripe_subscription_id FROM users WHERE id = $1", id)

	var user models.User
	if err := res.Scan(&user.ID, &user.FirebaseID, &user.StripeID, &user.FirstName, &user.LastName, &user.Email, &user.WordGoal, &user.StripeSubscriptionID); err != nil {
		panic(err)
	}

	return &user, nil
}

func (r *queryResolver) UserByFirebaseID(ctx context.Context, firebaseID *string) (*models.User, error) {
	res := wrabitDB.LogAndQueryRow(r.db, "SELECT id, firebase_id, stripe_id, first_name, last_name, email, word_goal, stripe_subscription_id, created_at FROM users WHERE firebase_id = $1", firebaseID)

	var user models.User
	if err := res.Scan(&user.ID, &user.FirebaseID, &user.StripeID, &user.FirstName, &user.LastName, &user.Email, &user.WordGoal, &user.StripeSubscriptionID, &user.CreatedAt); err != nil {
		panic(err)
	}

	return &user, nil
}

func (r *queryResolver) Editors(ctx context.Context, id *string) ([]*models.Editor, error) {
	if user := auth.ForContext(ctx); user == nil {
		return []*models.Editor{}, fmt.Errorf("Access denied")
	}

	var editors []*models.Editor

	if id == nil {
		res := wrabitDB.LogAndQuery(r.db, "SELECT * FROM editors")
		defer res.Close()
		for res.Next() {
			var editor = new(models.Editor)
			if err := res.Scan(&editor.ID, &editor.UserID, &editor.ShowCounter, &editor.ShowPrompt, &editor.ShowCounter, &editor.CreatedAt, &editor.UpdatedAt); err != nil {
				panic(err)
			}

			editors = append(editors, editor)
		}
	} else {
		res := wrabitDB.LogAndQueryRow(r.db, "SELECT * FROM editors WHERE id = $1", id)

		var editor = new(models.Editor)
		if err := res.Scan(&editor.ID, &editor.UserID, &editor.ShowCounter, &editor.ShowPrompt, &editor.ShowCounter, &editor.CreatedAt, &editor.UpdatedAt); err != nil {
			panic(err)
		}

		editors = append(editors, editor)
	}

	return editors, nil
}

func (r *queryResolver) WordGoal(ctx context.Context, userID string, date string) (int, error) {
	ctxUser := auth.ForContext(ctx)
	if ctxUser == nil || ctxUser.Subject != userID {
		return 0, fmt.Errorf("Access denied")
	}

	res := wrabitDB.LogAndQueryRow(r.db, "SELECT word_goal FROM users WHERE firebase_id = $1", userID)
	var user models.User
	if err := res.Scan(&user.WordGoal); err != nil {
		panic(err)
	}

	// Multiplier starts at 10%
	multiplier := 0.1

	lastStreakDayCount := 0
	lastEntryID := 0

	daySinceLastWrote := 0
	entryID := 0

	// Get last streak
	res = wrabitDB.LogAndQueryRow(r.db, "SELECT day_count, last_entry_id FROM streaks WHERE user_id = $1 ORDER BY updated_at DESC LIMIT 1;", userID)
	err := res.Scan(&lastStreakDayCount, &lastEntryID)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	// Figure out when the last entry was written
	// Is 0 if they wrote within 24 hours
	res = wrabitDB.LogAndQueryRow(r.db, "SELECT date_part('day', $1 - created_at::timestamp) As day_since_last_entry, id FROM entries WHERE user_id = $2 AND goal_hit = true ORDER BY created_at DESC LIMIT 1;", date, userID)
	err = res.Scan(&daySinceLastWrote, &entryID)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	// TODO: Decrement from goal count instead of resetting to 0.1 if lastStreakDayCount < 10
	if daySinceLastWrote < 2 && lastStreakDayCount > 0 && lastStreakDayCount < 10 {
		multiplier = float64(lastStreakDayCount) / 10
		if lastEntryID == entryID {
			multiplier = multiplier + 0.1
		}
	} else if daySinceLastWrote < 2 && lastStreakDayCount >= 10 {
		multiplier = 1.0
	} else if lastStreakDayCount >= 10 && daySinceLastWrote > 0 && daySinceLastWrote < 10 {
		// 1 --> 0.9
		// 2 --> 0.8
		// 3 --> 0.7
		// ...
		multiplier = 1.0 - (float64(daySinceLastWrote) * 0.1)
	}

	// int truncates the float which is fine for my purposes
	wordGoal := int(float64(user.WordGoal) * multiplier)

	return wordGoal, nil
}

func (r *queryResolver) Stats(ctx context.Context, global bool) (*models.Stats, error) {
	user := auth.ForContext(ctx)
	if user == nil {
		return &models.Stats{}, fmt.Errorf("Access denied")
	}

	var stats = new(models.Stats)

	var res = new(sql.Row)
	wordsWrittenQuery := "SELECT sum(word_count) as words_written FROM entries"
	if global {
		res = wrabitDB.LogAndQueryRow(r.db, wordsWrittenQuery)
	} else {
		wordsWrittenQuery = wordsWrittenQuery + " WHERE user_id = $1"
		res = wrabitDB.LogAndQueryRow(r.db, wordsWrittenQuery, user.Subject)
	}
	err := res.Scan(&stats.WordsWritten)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	longestEntryQuery := "SELECT max(word_count) as longest_entry FROM entries"
	if global {
		res = wrabitDB.LogAndQueryRow(r.db, longestEntryQuery)
	} else {
		longestEntryQuery = longestEntryQuery + " WHERE user_id = $1"
		res = wrabitDB.LogAndQueryRow(r.db, longestEntryQuery, user.Subject)
	}
	err = res.Scan(&stats.LongestEntry)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	longestStreakQuery := "SELECT max(day_count) as longest_streak FROM streaks"
	if global {
		res = wrabitDB.LogAndQueryRow(r.db, longestStreakQuery)
	} else {
		longestStreakQuery = longestStreakQuery + " WHERE user_id = $1"
		res = wrabitDB.LogAndQueryRow(r.db, longestStreakQuery, user.Subject)
	}
	err = res.Scan(&stats.LongestStreak)
	// Check if there hasn't been a streak yet
	if err != nil && err.Error() == "sql: Scan error on column index 0, name \"longest_streak\": converting NULL to int is unsupported" {
		stats.LongestStreak = 0
	} else if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	preferredDayOfWeekQuery := "SELECT preferred_day_of_week FROM (SELECT date_part('dow', updated_at) as preferred_day_of_week FROM entries) sub GROUP BY 1 ORDER BY count(*) DESC LIMIT 1"
	if global {
		res = wrabitDB.LogAndQueryRow(r.db, preferredDayOfWeekQuery)
	} else {
		preferredDayOfWeekQuery = "SELECT preferred_day_of_week FROM (SELECT date_part('dow', updated_at) as preferred_day_of_week FROM entries WHERE user_id = $1) sub GROUP BY 1 ORDER BY count(*) DESC LIMIT 1"
		res = wrabitDB.LogAndQueryRow(r.db, preferredDayOfWeekQuery, user.Subject)
	}
	err = res.Scan(&stats.PreferredDayOfWeek)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	var resTimes = new(sql.Rows)
	preferredWritingTimesQuery := "SELECT hour, count(*) FROM (SELECT date_part('hour', updated_at) as hour FROM entries) sub GROUP BY 1 ORDER BY 2 DESC"
	if global {
		resTimes = wrabitDB.LogAndQuery(r.db, preferredWritingTimesQuery)
	} else {
		preferredWritingTimesQuery := "SELECT hour, count(*) FROM (SELECT date_part('hour', updated_at) as hour FROM entries WHERE user_id = $1) sub GROUP BY 1 ORDER BY 2 DESC"
		resTimes = wrabitDB.LogAndQuery(r.db, preferredWritingTimesQuery, user.Subject)
	}
	defer resTimes.Close()
	for resTimes.Next() {
		var preferredWritingTime = new(models.PreferredWritingTime)
		if err := resTimes.Scan(&preferredWritingTime.Hour, &preferredWritingTime.Count); err != nil {
			panic(err)
		}

		stats.PreferredWritingTimes = append(stats.PreferredWritingTimes, preferredWritingTime)
	}

	return stats, nil
}

// Individul resolvers
type editorResolver struct{ *Resolver }

func (r *editorResolver) User(ctx context.Context, obj *models.Editor) (*models.User, error) {
	if user := auth.ForContext(ctx); user == nil {
		return &models.User{}, fmt.Errorf("Access denied")
	}

	res := wrabitDB.LogAndQueryRow(r.db, "SELECT * FROM users WHERE firebase_id = $1", obj.UserID)

	var user models.User
	if err := res.Scan(&user.ID, &user.FirebaseID, &user.StripeID, &user.FirstName, &user.LastName, &user.Email, &user.WordGoal, &user.CreatedAt, &user.UpdatedAt); err != nil {
		panic(err)
	}

	return &user, nil
}

type streakResolver struct{ *Resolver }

func (r *streakResolver) User(ctx context.Context, obj *models.Streak) (*models.User, error) {
	if user := auth.ForContext(ctx); user == nil {
		return &models.User{}, fmt.Errorf("Access denied")
	}

	res := wrabitDB.LogAndQueryRow(r.db, "SELECT * FROM users WHERE firebase_id = $1", obj.UserID)

	var user models.User
	if err := res.Scan(&user.ID, &user.FirebaseID, &user.StripeID, &user.FirstName, &user.LastName, &user.Email, &user.WordGoal, &user.CreatedAt, &user.UpdatedAt); err != nil {
		panic(err)
	}

	return &user, nil
}

func (r *streakResolver) LastEntryID(ctx context.Context, obj *models.Streak) (string, error) {
	return obj.LastEntryID, nil
}

type userResolver struct{ *Resolver }

func (r *userResolver) StripeSubscription(ctx context.Context, obj *models.User) (*models.StripeSubscription, error) {
	// Initialize Stripe
	stripe.Key = os.Getenv("STRIPE_KEY")

	if obj.StripeSubscriptionID == nil {
		return &models.StripeSubscription{}, nil
	}

	subscription, err := sub.Get(
		*obj.StripeSubscriptionID,
		nil,
	)
	if err != nil {
		return &models.StripeSubscription{}, nil
	}

	userSubscription := &models.StripeSubscription{
		ID:               subscription.ID,
		CurrentPeriodEnd: subscription.CurrentPeriodEnd,
		TrialEnd:         subscription.TrialEnd,
		CancelAt:         subscription.CancelAt,
		Status:           subscription.Status,
		Plan:             subscription.Plan,
	}

	return userSubscription, nil
}

type stripeSubscriptionResolver struct{ *Resolver }

func (r *stripeSubscriptionResolver) Plan(ctx context.Context, obj *models.StripeSubscription) (*models.Plan, error) {
	if obj.ID == "" {
		return &models.Plan{}, nil
	}

	plan := &models.Plan{
		ID:       obj.Plan.ID,
		Nickname: obj.Plan.Nickname,
		Product:  obj.Plan.Product.ID,
	}

	return plan, nil
}

func (r *stripeSubscriptionResolver) Status(ctx context.Context, obj *models.StripeSubscription) (string, error) {
	return fmt.Sprintf("%s", obj.Status), nil
}
