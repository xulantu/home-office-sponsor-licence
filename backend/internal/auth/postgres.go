package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"sponsor-tracker/internal/database"
)

// PostgresUserStore implements UserStore using PostgreSQL.
type PostgresUserStore struct {
	pool *pgxpool.Pool
}

func NewPostgresUserStore(pool *pgxpool.Pool) *PostgresUserStore {
	return &PostgresUserStore{pool: pool}
}

func (s *PostgresUserStore) FindUserByID(ctx context.Context, id int) (database.User, bool, error) {
	return database.FindUserByID(ctx, s.pool, id)
}

func (s *PostgresUserStore) FindUserByUsername(ctx context.Context, username string) (database.User, bool, error) {
	return database.FindUserByUsername(ctx, s.pool, username)
}

// PostgresSessionStore implements SessionStore using PostgreSQL.
type PostgresSessionStore struct {
	pool *pgxpool.Pool
}

func NewPostgresSessionStore(pool *pgxpool.Pool) *PostgresSessionStore {
	return &PostgresSessionStore{pool: pool}
}

func (s *PostgresSessionStore) CreateSession(ctx context.Context, userID int, expiry time.Duration) (string, error) {
	return database.CreateSession(ctx, s.pool, userID, expiry)
}

func (s *PostgresSessionStore) FindSession(ctx context.Context, token string) (database.Session, bool, error) {
	return database.FindSession(ctx, s.pool, token)
}

func (s *PostgresSessionStore) DeleteSession(ctx context.Context, token string) error {
	return database.DeleteSession(ctx, s.pool, token)
}

func (s *PostgresSessionStore) ExtendSession(ctx context.Context, token string, expiry time.Duration) error {
	return database.ExtendSession(ctx, s.pool, token, expiry)
}
