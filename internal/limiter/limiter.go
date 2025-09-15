package limiter

import (
    "net"
    "net/http"
    "strings"
    "time"

    "github.com/fabiuhp/rate-limiter-pos/internal/config"
    "github.com/fabiuhp/rate-limiter-pos/internal/storage"
)

type Limiter struct {
    cfg   config.Config
    store storage.Store
}

func NewLimiter(cfg config.Config, store storage.Store) *Limiter {
    return &Limiter{cfg: cfg, store: store}
}

type Decision struct {
    Allowed    bool
    RetryAfter time.Duration
    Remaining  int
}

func (l *Limiter) Evaluate(r *http.Request) (Decision, error) {
    token := strings.TrimSpace(r.Header.Get(l.cfg.APIKeyHeader))
    window := time.Second

    useToken := l.cfg.Strategy == "token" || l.cfg.Strategy == "both"
    useIP := l.cfg.Strategy == "ip" || l.cfg.Strategy == "both"
    if useToken && token != "" {
        if rule, ok := l.cfg.TokenRules[token]; ok {
            res, err := l.store.Attempt("token", token, rule.LimitPerSecond, window, rule.BlockFor)
            if err != nil { return Decision{}, err }
            return Decision{Allowed: res.Allowed, RetryAfter: res.RetryAfter, Remaining: res.Remaining}, nil
        }
        if l.cfg.Strategy == "token" && l.cfg.TokenDefaultLimitPerSecond > 0 {
            res, err := l.store.Attempt("token", token, l.cfg.TokenDefaultLimitPerSecond, window, l.cfg.TokenDefaultBlockFor)
            if err != nil { return Decision{}, err }
            return Decision{Allowed: res.Allowed, RetryAfter: res.RetryAfter, Remaining: res.Remaining}, nil
        }
    }

    if useIP {
        ip := clientIP(r)
        res, err := l.store.Attempt("ip", ip, l.cfg.IPLimitPerSecond, window, l.cfg.IPBlockFor)
        if err != nil { return Decision{}, err }
        return Decision{Allowed: res.Allowed, RetryAfter: res.RetryAfter, Remaining: res.Remaining}, nil
    }
    return Decision{Allowed: true, Remaining: 0}, nil
}

func clientIP(r *http.Request) string {
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        parts := strings.Split(xff, ",")
        if len(parts) > 0 {
            return strings.TrimSpace(parts[0])
        }
    }
    if xr := r.Header.Get("X-Real-IP"); xr != "" {
        return strings.TrimSpace(xr)
    }
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        return r.RemoteAddr
    }
    return host
}
