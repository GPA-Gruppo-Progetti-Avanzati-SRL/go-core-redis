// Package locker provides a Redis-backed implementation of the neutral
// go-core-app/lock.Locker primitive, built directly on redsync (Redlock). It is
// backend-agnostic toward its consumers: go-core-batch adapts a lock.Locker to
// gocron, but this package does not depend on gocron.
package locker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-core-app/lock"
	"github.com/go-redsync/redsync/v4"
	redsyncgoredis "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredis "github.com/redis/go-redis/v9"
)

// defaultExpiry is the redsync mutex TTL. The lock is a dispatch-dedup
// optimization (correctness lives elsewhere, e.g. DB claiming), so a single
// acquisition attempt with a modest explicit expiry is intentional: if the
// holder outlives the TTL the lock simply lapses instead of blocking forever.
const defaultExpiry = 30 * time.Second

type redisLocker struct {
	rs     *redsync.Redsync
	expiry time.Duration
}

// New returns a Redis-backed lock.Locker over the given client.
func New(client *goredis.Client) lock.Locker {
	pool := redsyncgoredis.NewPool(client)
	return &redisLocker{rs: redsync.New(pool), expiry: defaultExpiry}
}

// Acquire honours the neutral AcquireOption set: Tries/RetryDelay make redsync
// block and retry on contention (default: a single, non-blocking attempt), and
// Expiry overrides the mutex TTL (default: defaultExpiry). redsync natively
// implements the retry-until-acquired loop.
func (l *redisLocker) Acquire(ctx context.Context, key string, opts ...lock.AcquireOption) (lock.Handle, error) {
	cfg := lock.ResolveAcquireConfig(opts...)

	tries := cfg.Tries
	if tries < 1 {
		tries = 1
	}
	expiry := cfg.Expiry
	if expiry <= 0 {
		expiry = l.expiry
	}
	muOpts := []redsync.Option{redsync.WithTries(tries), redsync.WithExpiry(expiry)}
	if cfg.RetryDelay > 0 {
		muOpts = append(muOpts, redsync.WithRetryDelay(cfg.RetryDelay))
	}

	mu := l.rs.NewMutex(key, muOpts...)
	if err := mu.LockContext(ctx); err != nil {
		// Contention (quorum already taken) → not acquired; anything else is a
		// backend failure surfaced to the caller.
		var taken *redsync.ErrTaken
		if errors.As(err, &taken) || errors.Is(err, redsync.ErrFailed) {
			return nil, lock.ErrNotAcquired
		}
		return nil, fmt.Errorf("redis lock acquire %q: %w", key, err)
	}
	return &redisHandle{mu: mu}, nil
}

type redisHandle struct {
	mu *redsync.Mutex
}

func (h *redisHandle) Release(ctx context.Context) error {
	ok, err := h.mu.UnlockContext(ctx)
	if err != nil {
		// The lock having already expired is benign for a dispatch-dedup lock.
		if errors.Is(err, redsync.ErrLockAlreadyExpired) {
			return nil
		}
		return fmt.Errorf("redis lock release %q: %w", h.mu.Name(), err)
	}
	if !ok {
		return fmt.Errorf("redis lock release %q: not held", h.mu.Name())
	}
	return nil
}

// Extend renews the mutex TTL. Unlike Release, a lapsed lock here is not benign:
// the critical section believed it still held the lock, so a failed extension is
// surfaced as lock.ErrLockLost.
func (h *redisHandle) Extend(ctx context.Context) error {
	ok, err := h.mu.ExtendContext(ctx)
	if err != nil {
		if errors.Is(err, redsync.ErrLockAlreadyExpired) {
			return fmt.Errorf("redis lock extend %q: %w", h.mu.Name(), lock.ErrLockLost)
		}
		return fmt.Errorf("redis lock extend %q: %w", h.mu.Name(), err)
	}
	if !ok {
		return fmt.Errorf("redis lock extend %q: %w", h.mu.Name(), lock.ErrLockLost)
	}
	return nil
}
