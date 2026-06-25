package idempotency

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/kabirnarang39/stripe-billing/internal/cache"
)

const (
	lockTTL    = 30 * time.Second
	resultTTL  = 24 * time.Hour
	lockSuffix = ":lock"
)

// Store provides idempotency tracking backed by Redis.
// It uses a two-key approach: a "lock" key acquired via SetNX and a "result"
// key (the idempotency key itself) populated after the operation completes.
type Store struct {
	cache *cache.Client
}

// New creates a new idempotency Store.
func New(c *cache.Client) *Store {
	return &Store{cache: c}
}

// Begin checks for an existing cached response for the given userID and key.
//
// Return semantics:
//   - (cached, false, nil): a completed response was found; caller should return it.
//   - (nil, true, nil): no cached response; the lock was acquired; caller is the leader.
//   - (nil, false, nil): a lock exists but no result yet; concurrent in-flight request.
func (s *Store) Begin(ctx context.Context, userID, key string) (cached []byte, isLeader bool, err error) {
	resultKey := cache.KeyIdempotency(userID, key)
	lockKey := resultKey + lockSuffix

	// Check for an already-completed response.
	val, err := s.cache.Get(ctx, resultKey)
	if err == nil {
		// Found a completed response.
		return []byte(val), false, nil
	}
	if err != redis.Nil {
		// Redis error — fail open (treat as no cache).
		return nil, true, nil
	}

	// No result yet. Try to acquire the lock.
	acquired, err := s.cache.SetNX(ctx, lockKey, "1", lockTTL)
	if err != nil {
		// Redis error on lock — fail open; let caller proceed.
		return nil, true, nil
	}
	if !acquired {
		// Another request holds the lock; caller should retry/wait.
		return nil, false, nil
	}

	// Lock acquired; caller is the leader for this idempotency key.
	return nil, true, nil
}

// Complete stores the operation result under the idempotency key with a 24-hour TTL
// and removes the lock key.
func (s *Store) Complete(ctx context.Context, userID, key string, response []byte) error {
	resultKey := cache.KeyIdempotency(userID, key)
	lockKey := resultKey + lockSuffix

	if err := s.cache.Set(ctx, resultKey, string(response), resultTTL); err != nil {
		return err
	}
	// Best-effort lock removal; the TTL will clean it up if this fails.
	_ = s.cache.Del(ctx, lockKey)
	return nil
}
