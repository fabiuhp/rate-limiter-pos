package limiter_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/fabiuhp/rate-limiter-pos/internal/config"
    "github.com/fabiuhp/rate-limiter-pos/internal/limiter"
    "github.com/fabiuhp/rate-limiter-pos/internal/middleware"
    "github.com/fabiuhp/rate-limiter-pos/internal/storage"
)

func TestLimiter_IPMode(t *testing.T) {
    store := storage.NewMemoryStore()
    cfg := config.Config{
        Strategy:          "ip",
        APIKeyHeader:      "API_KEY",
        IPLimitPerSecond:  2,
        IPBlockFor:        300 * time.Millisecond,
        StoreDriver:       "memory",
    }
    l := limiter.NewLimiter(cfg, store)

    h := middleware.RateLimitMiddleware(l)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200); _, _ = w.Write([]byte("ok")) }))

    rr := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/", nil)
    req.RemoteAddr = "1.2.3.4:12345"

    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("want 200, got %d", rr.Code) }
    rr = httptest.NewRecorder()
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("want 200, got %d", rr.Code) }
    rr = httptest.NewRecorder()
    h.ServeHTTP(rr, req)
    if rr.Code != 429 { t.Fatalf("want 429, got %d", rr.Code) }
}

func TestLimiter_TokenOverridesIP(t *testing.T) {
    store := storage.NewMemoryStore()
    cfg := config.Config{
        Strategy:          "both",
        APIKeyHeader:      "API_KEY",
        IPLimitPerSecond:  1,
        IPBlockFor:        300 * time.Millisecond,
        TokenRules: map[string]config.TokenRule{
            "abc123": {LimitPerSecond: 3, BlockFor: 300 * time.Millisecond},
        },
        StoreDriver:       "memory",
    }
    l := limiter.NewLimiter(cfg, store)
    h := middleware.RateLimitMiddleware(l)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))

    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("API_KEY", "abc123")
    req.RemoteAddr = "1.2.3.4:12345"

    for i := range 3 {
        rr := httptest.NewRecorder()
        h.ServeHTTP(rr, req)
        if rr.Code != 200 { t.Fatalf("want 200 @%d, got %d", i, rr.Code) }
    }
    rr := httptest.NewRecorder()
    h.ServeHTTP(rr, req)
    if rr.Code != 429 { t.Fatalf("want 429 after token limit, got %d", rr.Code) }
}
