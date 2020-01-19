package db

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/writewithwrabit/server/models"
)

func TestLogAndQueryShouldReturnResults(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows(
		[]string{"id", "firebase_id", "stripe_id", "stripe_subscription_id", "first_name", "last_name", "email", "word_goal"},
	).
		AddRow(1, "IWoB2L4lcJW8brqOHd7oJfzn8vt2", "cus_GIHI1V0ryeznB2", "sub_GIHImr4be4B275", "Test", "Account", "testing@writewithwrabit.com", 1000)

	mock.ExpectQuery("SELECT id, firebase_id, stripie_id, firstName, lastName, email, word_goal FROM users").WillReturnRows(rows)

	res := LogAndQuery(db, "SELECT id, firebase_id, stripie_id, firstName, lastName, email, word_goal FROM users")
	var user models.User
	err = res.Scan(&user.ID, &user.FirebaseID, &user.StripeID, &user.FirstName, &user.LastName, &user.Email, &user.WordGoal)

	assert.NotEmpty(t, res)
	assert.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestLogAndQueryRowShouldReturnResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows(
		[]string{"id", "firebase_id", "stripe_id", "stripe_subscription_id", "first_name", "last_name", "email", "word_goal"},
	).AddRow(1, "IWoB2L4lcJW8brqOHd7oJfzn8vt2", "cus_GIHI1V0ryeznB2", "sub_GIHImr4be4B275", "Test", "Account", "testing@writewithwrabit.com", 1000)

	mock.ExpectQuery("SELECT id, firebase_id, stripie_id, firstName, lastName, email, word_goal FROM users").WillReturnRows(rows)

	res := LogAndQueryRow(db, "SELECT id, firebase_id, stripie_id, firstName, lastName, email, word_goal FROM users")
	var user models.User
	err = res.Scan(&user.ID, &user.FirebaseID, &user.StripeID, &user.FirstName, &user.LastName, &user.Email, &user.WordGoal)

	assert.NotEmpty(t, res)
	assert.NotNil(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
