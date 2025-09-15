package storage

import (
    "time"
)

type AttemptResult struct {
    Allowed     bool
    Remaining   int
    RetryAfter  time.Duration
    WindowReset time.Time
}

type Store interface {
    Attempt(scope, key string, limit int, window time.Duration, blockFor time.Duration) (AttemptResult, error)
    Close() error
}
