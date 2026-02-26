package database

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
)

// InvitationCode represents a registration invitation code.
type InvitationCode struct {
	ID        int
	Code      string
	Count     int
	CreatedAt time.Time
	ExpiresAt time.Time
}

// FindInvitationCode returns the most recently inserted, non-expired invitation
// code row matching the given code string.
// Returns the code and true if found, or empty and false if not found.
func FindInvitationCode(ctx context.Context, q Querier, code string) (InvitationCode, bool, error) {
	var ic InvitationCode
	err := q.QueryRow(ctx,
		`SELECT id, code, count, created_at, expires_at
		 FROM invitation_codes
		 WHERE code = $1 AND expires_at > NOW()
		 ORDER BY id DESC
		 LIMIT 1`,
		code,
	).Scan(&ic.ID, &ic.Code, &ic.Count, &ic.CreatedAt, &ic.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return InvitationCode{}, false, nil
	}
	if err != nil {
		return InvitationCode{}, false, fmt.Errorf("find invitation code: %w", err)
	}
	return ic, true, nil
}

// CountInvitationCodeUses returns the number of users registered with the
// given invitation code ID. Returns math.MaxInt on error so that a forgotten
// error check fails closed (blocks registration) rather than open.
func CountInvitationCodeUses(ctx context.Context, q Querier, codeID int) (int, error) {
	var n int
	err := q.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE invitation_code_id = $1`,
		codeID,
	).Scan(&n)
	if err != nil {
		return math.MaxInt, fmt.Errorf("count invitation code uses: %w", err)
	}
	return n, nil
}
