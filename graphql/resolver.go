//go:generate go run github.com/99designs/gqlgen

package server

import (
	"context"
	"fmt"
	"math/rand"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	editors []*Editor
	entries []*Entry
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
		ID:        fmt.Sprintf("T%d", rand.New(rand.NewSource(666)).Int31()),
		UserID:    input.UserID,
		Content:   input.Content,
		WordCount: input.WordCount,
	}

	r.entries = append(r.entries, entry)
	return entry, nil
}

func (r *mutationResolver) CreateEditor(ctx context.Context, input NewEditor) (*Editor, error) {
	editor := &Editor{
		ID:          fmt.Sprintf("T%d", rand.New(rand.NewSource(666)).Int31()),
		UserID:      input.UserID,
		ShowToolbar: input.ShowToolbar,
		ShowPrompt:  input.ShowPrompt,
		ShowCounter: input.ShowCounter,
	}

	r.editors = append(r.editors, editor)
	return editor, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Editors(ctx context.Context) ([]*Editor, error) {
	return r.editors, nil
}
func (r *queryResolver) Entries(ctx context.Context) ([]*Entry, error) {
	return r.entries, nil
}

type editorResolver struct{ *Resolver }

func (r *editorResolver) User(ctx context.Context, obj *Editor) (*User, error) {
	return &User{ID: obj.UserID}, nil
}

type entryResolver struct{ *Resolver }

func (r *entryResolver) User(ctx context.Context, obj *Entry) (*User, error) {
	return &User{ID: obj.UserID}, nil
}
