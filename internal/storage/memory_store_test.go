package storage

import (
    "testing"
    "time"
)

func TestMemoryStore_AllowsWithinLimit(t *testing.T) {
    m := NewMemoryStore()
    defer m.Close()
    window := time.Second
    block := 2 * time.Second
    for i := range 5 {
        res, err := m.Attempt("ip", "1.2.3.4", 5, window, block)
        if err != nil { t.Fatal(err) }
        if !res.Allowed {
            t.Fatalf("expected allowed on request %d", i+1)
        }
    }
}

func TestMemoryStore_BlocksAfterExceed(t *testing.T) {
    m := NewMemoryStore()
    defer m.Close()
    window := 200 * time.Millisecond
    block := 300 * time.Millisecond
    for range 3 { _, _ = m.Attempt("ip", "1.2.3.5", 2, window, block) }
    res, err := m.Attempt("ip", "1.2.3.5", 2, window, block)
    if err != nil { t.Fatal(err) }
    if res.Allowed {
        t.Fatalf("expected blocked after exceeding limit")
    }
    time.Sleep(block + 50*time.Millisecond)
    res, err = m.Attempt("ip", "1.2.3.5", 2, window, block)
    if err != nil { t.Fatal(err) }
    if !res.Allowed {
        t.Fatalf("expected allowed after block expiry")
    }
}
