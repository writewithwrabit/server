package resolvers

import (
	"context"
	"testing"

	firebase "firebase.google.com/go/auth"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/writewithwrabit/server/auth"
)

func TestDeleteEntryWithoutUser(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	resolver := &Resolver{
		db: db,
	}
	mutResolver := &mutationResolver{
		Resolver: resolver,
	}

	c := context.Background()

	res, err := mutResolver.DeleteEntry(c, "1")

	assert.Empty(t, res)
	assert.NotEmpty(t, err)
	assert.Equal(t, "Access denied", err.Error())
}

func TestDeleteEntry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	resolver := &Resolver{
		db: db,
	}
	mutResolver := &mutationResolver{
		Resolver: resolver,
	}

	token := &firebase.Token{
		Subject: "abcdefg",
	}

	c := context.Background()
	ctx := context.WithValue(c, auth.UserCtxKey, token)

	result := sqlmock.NewResult(1, 1)
	mock.ExpectExec("DELETE FROM entries WHERE user_id \\= \\$1 AND id \\= \\$2").
		WithArgs("abcdefg", "1").WillReturnResult(result)

	res, err := mutResolver.DeleteEntry(ctx, "1")

	assert.Equal(t, res.ID, "1")
	assert.Empty(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
