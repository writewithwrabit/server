package resolvers

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/sub"
	"github.com/writewithwrabit/server/auth"
	cryptopasta "github.com/writewithwrabit/server/cryptopasta"
	wrabitDB "github.com/writewithwrabit/server/db"
	"github.com/writewithwrabit/server/models"
)

func (r *queryResolver) Entries(ctx context.Context, id *string) ([]*models.Entry, error) {
	if user := auth.ForContext(ctx); user == nil {
		return []*models.Entry{}, fmt.Errorf("Access denied")
	}

	key := [32]byte{}
	keyString := os.Getenv("ENCRYPTION_KEY")
	copy(key[:], keyString)

	var entries []*models.Entry

	if id == nil {
		res := wrabitDB.LogAndQuery(r.db, "SELECT * FROM entries")
		defer res.Close()
		for res.Next() {
			var entry = new(models.Entry)
			if err := res.Scan(&entry.ID, &entry.UserID, &entry.WordCount, &entry.Content, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
				panic(err)
			}

			decodedContent, err := hex.DecodeString(entry.Content)
			content, err := cryptopasta.Decrypt(decodedContent, &key)
			if err == nil {
				entry.Content = string(content)
			}

			entries = append(entries, entry)
		}
	} else {
		res := wrabitDB.LogAndQueryRow(r.db, "SELECT * FROM entries WHERE id = $1", id)

		var entry = new(models.Entry)
		if err := res.Scan(&entry.ID, &entry.UserID, &entry.WordCount, &entry.Content, &entry.CreatedAt, &entry.UpdatedAt, &entry.GoalHit); err != nil {
			panic(err)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *queryResolver) EntriesByUserID(ctx context.Context, userID string, startDate *string, endDate *string) ([]*models.Entry, error) {
	if user := auth.ForContext(ctx); user == nil {
		return []*models.Entry{}, fmt.Errorf("Access denied")
	}

	key := [32]byte{}
	keyString := os.Getenv("ENCRYPTION_KEY")
	copy(key[:], keyString)

	var entries []*models.Entry

	var res *sql.Rows
	if startDate != nil && endDate == nil {
		res = wrabitDB.LogAndQuery(r.db, "SELECT * FROM entries WHERE user_id = $1 AND created_at >= $2 AND word_count > 0 ORDER BY created_at DESC", userID, startDate)
	} else if startDate == nil && endDate != nil {
		res = wrabitDB.LogAndQuery(r.db, "SELECT * FROM entries WHERE user_id = $1 AND created_at <= $2 AND word_count > 0 ORDER BY created_at DESC", userID, endDate)
	} else if startDate != nil && endDate != nil {
		res = wrabitDB.LogAndQuery(r.db, "SELECT * FROM entries WHERE user_id = $1 AND created_at >= $2 AND word_count > 0 AND created_at <= $3 ORDER BY created_at DESC", userID, startDate, endDate)
	} else {
		res = wrabitDB.LogAndQuery(r.db, "SELECT * FROM entries WHERE user_id = $1 AND word_count > 0 ORDER BY created_at DESC", userID)
	}

	defer res.Close()
	for res.Next() {
		var entry = new(models.Entry)
		if err := res.Scan(&entry.ID, &entry.UserID, &entry.WordCount, &entry.Content, &entry.CreatedAt, &entry.UpdatedAt, &entry.GoalHit); err != nil {
			panic(err)
		}

		decodedContent, err := hex.DecodeString(entry.Content)
		content, err := cryptopasta.Decrypt(decodedContent, &key)
		if err == nil {
			entry.Content = string(content)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *queryResolver) DailyEntry(ctx context.Context, userID string, date string) (*models.Entry, error) {
	if user := auth.ForContext(ctx); user == nil {
		return &models.Entry{}, fmt.Errorf("Access denied")
	}

	key := [32]byte{}
	keyString := os.Getenv("ENCRYPTION_KEY")
	copy(key[:], keyString)

	res := wrabitDB.LogAndQueryRow(r.db, "SELECT * FROM entries WHERE user_id = $1 AND created_at >= $2 ORDER BY created_at DESC", userID, date)

	var entry = new(models.Entry)
	err := res.Scan(&entry.ID, &entry.UserID, &entry.WordCount, &entry.Content, &entry.CreatedAt, &entry.UpdatedAt, &entry.GoalHit)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	if err == sql.ErrNoRows {
		res := wrabitDB.LogAndQueryRow(r.db, "INSERT INTO entries (user_id, content, word_count, created_at) VALUES ($1, $2, $3, $4) RETURNING id", userID, "", 0, date)
		if err := res.Scan(&entry.ID); err != nil {
			panic(err)
		}
	} else {
		decodedContent, err := hex.DecodeString(entry.Content)
		content, err := cryptopasta.Decrypt(decodedContent, &key)
		if err == nil {
			entry.Content = string(content)
		}
	}

	return entry, nil
}

func (r *mutationResolver) CreateEntry(ctx context.Context, input models.NewEntry) (*models.Entry, error) {
	if user := auth.ForContext(ctx); user == nil {
		return &models.Entry{}, fmt.Errorf("Access denied")
	}

	entry := &models.Entry{
		ID:        "",
		UserID:    input.UserID,
		Content:   input.Content,
		WordCount: input.WordCount,
	}

	res := wrabitDB.LogAndQueryRow(r.db, "INSERT INTO entries (user_id, content, word_count) VALUES ($1, $2, $3) RETURNING id", entry.UserID, entry.Content, entry.WordCount)
	if err := res.Scan(&entry.ID); err != nil {
		panic(err)
	}

	return entry, nil
}

func (r *mutationResolver) UpdateEntry(ctx context.Context, id string, input models.ExistingEntry, date string) (*models.Entry, error) {
	user := auth.ForContext(ctx)
	if user == nil || user.Subject != input.UserID {
		return &models.Entry{}, fmt.Errorf("Access denied")
	}

	key := [32]byte{}
	keyString := os.Getenv("ENCRYPTION_KEY")
	copy(key[:], keyString)

	// Encrypt the content for the database
	// but return the unencrypted content to the client
	content, err := cryptopasta.Encrypt([]byte(input.Content), &key)
	if err != nil {
		panic(err)
	}

	entry := &models.Entry{
		ID:        id,
		UserID:    input.UserID,
		Content:   input.Content,
		WordCount: input.WordCount,
		GoalHit:   input.GoalHit,
	}

	res := wrabitDB.LogAndQueryRow(r.db, "UPDATE entries SET content = $1, word_count = $2, goal_hit = $3 WHERE id = $4 AND user_id = $5 RETURNING id", hex.EncodeToString(content), entry.WordCount, entry.GoalHit, entry.ID, entry.UserID)
	if err := res.Scan(&entry.ID); err != nil {
		panic(err)
	}

	// TODO: Move this logic into a sane place
	if input.GoalHit {
		// Get the latest streak for the user
		res := wrabitDB.LogAndQueryRow(r.db, "SELECT ID, user_id, day_count, last_entry_id FROM streaks WHERE user_id = $1 AND updated_at >= $2::timestamp - INTERVAL '1 DAY' ORDER BY created_at DESC LIMIT 1", entry.UserID, date)

		var streak = new(models.Streak)
		err := res.Scan(&streak.ID, &streak.UserID, &streak.DayCount, &streak.LastEntryID)
		if err != nil && err != sql.ErrNoRows {
			panic(err)
		}
		newStreakCount := streak.DayCount + 1

		// If no streak exists, create one
		if err == sql.ErrNoRows {
			res := wrabitDB.LogAndQueryRow(r.db, "INSERT INTO streaks (user_id, day_count, last_entry_id) VALUES ($1, $2, $3) RETURNING id", entry.UserID, 1, entry.ID)
			if err := res.Scan(&entry.ID); err != nil {
				panic(err)
			}
		} else if streak != nil && streak.LastEntryID != entry.ID {
			res := wrabitDB.LogAndQueryRow(r.db, "UPDATE streaks SET last_entry_id = $1, day_count = $2 WHERE id = $3 AND user_id = $4 RETURNING id", entry.ID, newStreakCount, streak.ID, entry.UserID)
			if err := res.Scan(&entry.ID); err != nil {
				panic(err)
			}
		}

		// The streak is not valid for a donation
		// return early to save network/DB calls
		if newStreakCount%7 != 0 {
			return entry, nil
		}

		// Add donation if sequired
		res = wrabitDB.LogAndQueryRow(r.db, "SELECT stripe_subscription_id FROM users WHERE firebase_id = $1", entry.UserID)

		var user models.User
		if err := res.Scan(&user.StripeSubscriptionID); err != nil {
			fmt.Println(err)
			return entry, nil
		}

		// TODO: This logic is duplicated...
		stripe.Key = os.Getenv("STRIPE_KEY")

		subscription, err := sub.Get(
			*user.StripeSubscriptionID,
			nil,
		)
		if err != nil {
			fmt.Println(err)
			return entry, nil
		}

		userSubscription := &models.StripeSubscription{
			ID:               subscription.ID,
			CurrentPeriodEnd: subscription.CurrentPeriodEnd,
			TrialEnd:         subscription.TrialEnd,
			CancelAt:         subscription.CancelAt,
			Status:           subscription.Status,
			Plan:             subscription.Plan,
		}

		// Check subscription is valid
		// Status === active
		// CurrentPeriodEnd end has not passed
		if userSubscription.Status != "active" && !time.Now().Before(time.Unix(subscription.CurrentPeriodEnd, 0)) {
			return entry, nil
		}

		// Check to see if a donation has been made for the specific entry
		res = wrabitDB.LogAndQueryRow(r.db, "SELECT id FROM donations WHERE user_id = $1 AND entry_id = $2 LIMIT 1", entry.UserID, entry.ID)
		var donation = &models.Donation{}
		err = res.Scan(&donation.ID)
		if err != nil && err != sql.ErrNoRows {
			panic(err)
		}

		// No donation has been made
		if err == sql.ErrNoRows {
			res = wrabitDB.LogAndQueryRow(r.db, "INSERT INTO donations (user_id, amount, entry_id) VALUES ($1, $2, $3) RETURNING id", entry.UserID, 1, entry.ID)
			if err := res.Scan(&entry.ID); err != nil {
				fmt.Println(err)
				return entry, nil
			}
		}
	}

	return entry, nil
}

func (r *mutationResolver) DeleteEntry(ctx context.Context, id string) (*models.Entry, error) {
	user := auth.ForContext(ctx)
	fmt.Println(user)
	if user == nil {
		return &models.Entry{}, fmt.Errorf("Access denied")
	}

	var entry = &models.Entry{}
	res := wrabitDB.LogAndExec(r.db, "DELETE FROM entries WHERE user_id = $1 AND id = $2", user.Subject, id)
	count, err := res.RowsAffected()
	if err == nil && count == 1 {
		entry.ID = id
	}

	return entry, nil
}

type entryResolver struct{ *Resolver }

func (r *entryResolver) User(ctx context.Context, obj *models.Entry) (*models.User, error) {
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

func (r *entryResolver) GoalHit(ctx context.Context, obj *models.Entry) (bool, error) {
	return obj.GoalHit, nil
}
