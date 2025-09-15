package storage

import (
    "sync"
    "time"
)

type counter struct {
    count     int
    windowEnd time.Time
}

type ban struct {
    until time.Time
}

type MemoryStore struct {
    mu       sync.Mutex
    counters map[string]counter
    bans     map[string]ban
}

func NewMemoryStore() *MemoryStore {
    return &MemoryStore{
        counters: make(map[string]counter),
        bans:     make(map[string]ban),
    }
}

func (m *MemoryStore) Attempt(scope, key string, limit int, window time.Duration, blockFor time.Duration) (AttemptResult, error) {
    k := scope + ":" + key
    now := time.Now()

    m.mu.Lock()
    defer m.mu.Unlock()

    if b, ok := m.bans[k]; ok {
        if now.Before(b.until) {
            return AttemptResult{Allowed: false, Remaining: 0, RetryAfter: time.Until(b.until)}, nil
        }
        delete(m.bans, k)
    }

    c := m.counters[k]
    if now.After(c.windowEnd) {
        c = counter{count: 0, windowEnd: now.Add(window)}
    }
    c.count++
    remaining := limit - c.count
    allowed := c.count <= limit
    m.counters[k] = c

    if !allowed {
        until := now.Add(blockFor)
        m.bans[k] = ban{until: until}
        return AttemptResult{Allowed: false, Remaining: 0, RetryAfter: time.Until(until)}, nil
    }

    return AttemptResult{Allowed: true, Remaining: remaining, WindowReset: c.windowEnd}, nil
}

func (m *MemoryStore) Close() error { return nil }
