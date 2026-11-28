package store

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsInvalidUUID(t *testing.T) {
	if !isInvalidUUID(&pgconn.PgError{Code: "22P02"}) {
		t.Fatal("expected 22P02 to be treated as an invalid id")
	}

	if isInvalidUUID(errors.New("boom")) {
		t.Fatal("plain error must not be treated as an invalid id")
	}

	if isInvalidUUID(&pgconn.PgError{Code: "23505"}) {
		t.Fatal("other postgres errors must not be treated as an invalid id")
	}
}
