package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps a pgxpool.Pool and provides domain-level query methods.
type Store struct {
	pool *pgxpool.Pool
}

// New creates a new Store by connecting to the given PostgreSQL connection string.
func New(ctx context.Context, connStr string) (*Store, error) {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("store: open pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("store: ping: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Pool returns the underlying pgxpool.Pool for callers that need direct access.
func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// WithTx runs fn inside a serializable transaction. It commits on success and
// rolls back (and returns the original error) on failure.
func (s *Store) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("store: begin tx: %w", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("store: commit tx: %w", err)
	}
	return nil
}

// Close releases all connections in the pool.
func (s *Store) Close() {
	s.pool.Close()
}
