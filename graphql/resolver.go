//go:generate go run github.com/99designs/gqlgen

package server

import (
	"context"
  "database/sql"
  "fmt"

  wrabitDB "github.com/writewithwrabit/server/db"
)

type Resolver struct {
  db *sql.DB
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

func (r *mutationResolver) CreateEntry(ctx context.Context, input NewEntry) (*Entry, error) {
	entry := &Entry{
    ID: "",
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

func (r *mutationResolver) CreateEditor(ctx context.Context, input NewEditor) (*Editor, error) {
	editor := &Editor{
    ID: "",
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

func (r *queryResolver) Editors(ctx context.Context) ([]*Editor, error) {
	res := wrabitDB.LogAndQuery(r.db, "SELECT * FROM editors")
	defer res.Close()

	var editors []*Editor
	for res.Next() {
		var editor = new(Editor)
		if err := res.Scan(&editor.ID, &editor.UserID, &editor.ShowCounter, &editor.ShowPrompt, &editor.ShowCounter); err != nil {
			panic(err)
    }

		editors = append(editors, editor)
	}

	return editors, nil
}

func (r *queryResolver) Entries(ctx context.Context) ([]*Entry, error) {
	res := wrabitDB.LogAndQuery(r.db, "SELECT * FROM entries")
	defer res.Close()

	var entries []*Entry
	for res.Next() {
		var entry = new(Entry)
		if err := res.Scan(&entry.ID, &entry.UserID, &entry.WordCount, &entry.Content); err != nil {
			panic(err)
    }

		entries = append(entries, entry)
	}

	return entries, nil
}

type editorResolver struct{ *Resolver }

func (r *editorResolver) User(ctx context.Context, obj *Editor) (*User, error) {
  res := wrabitDB.LogAndQueryRow(r.db, "SELECT id, first_name, last_name, email, word_goal FROM users WHERE id = $1", obj.UserID)

	var user User
	if err := res.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.WordGoal); err != nil {
		panic(err)
  }

	return &user, nil
}

type entryResolver struct{ *Resolver }

func (r *entryResolver) User(ctx context.Context, obj *Entry) (*User, error) {
  res := wrabitDB.LogAndQueryRow(r.db, "SELECT id, first_name, last_name, email, word_goal FROM users WHERE id = $1", obj.UserID)

	var user User
	if err := res.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.WordGoal); err != nil {
		panic(err)
  }

	return &user, nil
}
