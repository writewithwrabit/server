//go:generate go run github.com/99designs/gqlgen

package server

import (
	"context"
	"database/sql"
	"fmt"

	wrabitDB "github.com/writewithwrabit/server/db"
)

type Resolver struct {
	db      *sql.DB
	editors []*Editor
	entries []*Entry
}

func New(db *sql.DB) Config {
	return Config{
		Resolvers: &Resolver{
			db: db,
		},
	}
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

func (r *Resolver) Editor() EditorResolver {
	return &editorResolver{r}
}

func (r *Resolver) Entry() EntryResolver {
	return &entryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateUser(ctx context.Context, input NewUser) (*User, error) {
	user := &User{
		FirebaseID: input.FirebaseID,
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		Email:      input.Email,
	}

	res := wrabitDB.LogAndQueryRow(r.db, "INSERT INTO users (firebase_id, first_name, last_name, email) VALUES ($1, $2, $3, $4) RETURNING id", user.FirebaseID, user.FirstName, user.LastName, user.Email)
	fmt.Println(res)
	if err := res.Scan(&user.ID); err != nil {
		panic(err)
	}

	return user, nil
}

func (r *mutationResolver) CreateEntry(ctx context.Context, input NewEntry) (*Entry, error) {
	entry := &Entry{
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

func (r *mutationResolver) UpdateEntry(ctx context.Context, id string, input ExistingEntry) (*Entry, error) {
	entry := &Entry{
		ID:        id,
		UserID:    input.UserID,
		Content:   input.Content,
		WordCount: input.WordCount,
	}

	res := wrabitDB.LogAndQueryRow(r.db, "UPDATE entries SET content = $1, word_count = $2 WHERE id = $3 AND user_id = $4 RETURNING id", entry.Content, entry.WordCount, entry.ID, entry.UserID)
	if err := res.Scan(&entry.ID); err != nil {
		panic(err)
	}

	return entry, nil
}

func (r *mutationResolver) CreateEditor(ctx context.Context, input NewEditor) (*Editor, error) {
	editor := &Editor{
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

func (r *queryResolver) User(ctx context.Context, firebaseID *string) (*User, error) {
	res := wrabitDB.LogAndQueryRow(r.db, "SELECT * FROM users WHERE firebase_id = $1", firebaseID)

	var user User
	if err := res.Scan(&user.ID, &user.FirebaseID, &user.FirstName, &user.LastName, &user.Email, &user.WordGoal); err != nil {
		panic(err)
	}

	return &user, nil
}

func (r *queryResolver) Editors(ctx context.Context, id *string) ([]*Editor, error) {
	var editors []*Editor

	if id == nil {
		res := wrabitDB.LogAndQuery(r.db, "SELECT * FROM editors")
		defer res.Close()
		for res.Next() {
			var editor = new(Editor)
			if err := res.Scan(&editor.ID, &editor.UserID, &editor.ShowCounter, &editor.ShowPrompt, &editor.ShowCounter, &editor.CreatedAt, &editor.UpdatedAt); err != nil {
				panic(err)
			}

			editors = append(editors, editor)
		}
	} else {
		res := wrabitDB.LogAndQueryRow(r.db, "SELECT * FROM editors WHERE id = $1", id)

		var editor = new(Editor)
		if err := res.Scan(&editor.ID, &editor.UserID, &editor.ShowCounter, &editor.ShowPrompt, &editor.ShowCounter, &editor.CreatedAt, &editor.UpdatedAt); err != nil {
			panic(err)
		}

		editors = append(editors, editor)
	}

	return editors, nil
}

func (r *queryResolver) Entries(ctx context.Context, id *string) ([]*Entry, error) {
	var entries []*Entry

	if id == nil {
		res := wrabitDB.LogAndQuery(r.db, "SELECT * FROM entries")
		defer res.Close()
		for res.Next() {
			var entry = new(Entry)
			if err := res.Scan(&entry.ID, &entry.UserID, &entry.WordCount, &entry.Content, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
				panic(err)
			}

			entries = append(entries, entry)
		}
	} else {
		res := wrabitDB.LogAndQueryRow(r.db, "SELECT * FROM entries WHERE id = $1", id)

		var entry = new(Entry)
		if err := res.Scan(&entry.ID, &entry.UserID, &entry.WordCount, &entry.Content, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
			panic(err)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *queryResolver) EntriesByUserID(ctx context.Context, userID string, startDate *string, endDate *string) ([]*Entry, error) {
	var entries []*Entry

	var res *sql.Rows
	if startDate != nil && endDate == nil {
		res = wrabitDB.LogAndQuery(r.db, "SELECT * FROM entries WHERE user_id = $1 AND created_at >= $2 ORDER BY created_at DESC", userID, startDate)
	} else if startDate == nil && endDate != nil {
		res = wrabitDB.LogAndQuery(r.db, "SELECT * FROM entries WHERE user_id = $1 AND created_at <= $2 ORDER BY created_at DESC", userID, endDate)
	} else if startDate != nil && endDate != nil {
		res = wrabitDB.LogAndQuery(r.db, "SELECT * FROM entries WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3 ORDER BY created_at DESC", userID, startDate, endDate)
	} else {
		res = wrabitDB.LogAndQuery(r.db, "SELECT * FROM entries WHERE user_id = $1 ORDER BY created_at DESC", userID)
	}

	defer res.Close()
	for res.Next() {
		var entry = new(Entry)
		if err := res.Scan(&entry.ID, &entry.UserID, &entry.WordCount, &entry.Content, &entry.CreatedAt, &entry.UpdatedAt); err != nil {
			panic(err)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *queryResolver) LatestEntry(ctx context.Context, userID string) (*Entry, error) {
	res := wrabitDB.LogAndQueryRow(r.db, "SELECT * FROM entries WHERE user_id = $1 AND created_at BETWEEN NOW() - INTERVAL '24 HOURS' AND NOW() ORDER BY created_at DESC", userID)

	fmt.Println(res)
	var entry = new(Entry)
	err := res.Scan(&entry.ID, &entry.UserID, &entry.WordCount, &entry.Content, &entry.CreatedAt, &entry.UpdatedAt)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	if err == sql.ErrNoRows {
		res := wrabitDB.LogAndQueryRow(r.db, "INSERT INTO entries (user_id, content, word_count) VALUES ($1, $2, $3) RETURNING id", userID, "", 0)
		if err := res.Scan(&entry.ID); err != nil {
			panic(err)
		}
	}

	return entry, nil
}

type editorResolver struct{ *Resolver }

func (r *editorResolver) User(ctx context.Context, obj *Editor) (*User, error) {
	res := wrabitDB.LogAndQueryRow(r.db, "SELECT * FROM users WHERE firebase_id = $1", obj.UserID)

	var user User
	if err := res.Scan(&user.ID, &user.FirebaseID, &user.FirstName, &user.LastName, &user.Email, &user.WordGoal, &user.CreatedAt, &user.UpdatedAt); err != nil {
		panic(err)
	}

	return &user, nil
}

type entryResolver struct{ *Resolver }

func (r *entryResolver) User(ctx context.Context, obj *Entry) (*User, error) {
	res := wrabitDB.LogAndQueryRow(r.db, "SELECT * FROM users WHERE firebase_id = $1", obj.UserID)

	var user User
	if err := res.Scan(&user.ID, &user.FirebaseID, &user.FirstName, &user.LastName, &user.Email, &user.WordGoal, &user.CreatedAt, &user.UpdatedAt); err != nil {
		panic(err)
	}

	return &user, nil
}
