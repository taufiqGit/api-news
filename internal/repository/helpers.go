package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrNotFound = errors.New("record not found")

// --- pgtype → *T helpers ---

func textToPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

func uuidToPtr(u pgtype.UUID) *uuid.UUID {
	if u.Valid {
		id := uuid.UUID(u.Bytes)
		return &id
	}
	return nil
}

func timestamptzToPtr(t pgtype.Timestamptz) *time.Time {
	if t.Valid {
		val := t.Time
		return &val
	}
	return nil
}

// --- *T → pgtype helpers ---

func ptrToText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func ptrToUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	var u pgtype.UUID
	u.Bytes = *id
	u.Valid = true
	return u
}

func ptrToTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func ptrToBool(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}
