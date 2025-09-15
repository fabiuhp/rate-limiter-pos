package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/fabiuhp/rate-limiter-pos/internal/config"
    "github.com/fabiuhp/rate-limiter-pos/internal/limiter"
    "github.com/fabiuhp/rate-limiter-pos/internal/middleware"
    "github.com/fabiuhp/rate-limiter-pos/internal/storage"
)

func main() {
    cfg := config.Load()

    var store storage.Store
    var err error
    switch cfg.StoreDriver {
    case "memory":
        store = storage.NewMemoryStore()
    default:
        store, err = storage.NewRedisStore(storage.RedisConfig{
            Addr:     cfg.Redis.Addr,
            Password: cfg.Redis.Password,
            DB:       cfg.Redis.DB,
        })
        if err != nil {
            log.Fatalf("failed to init redis store: %v", err)
        }
    }

    rl := limiter.NewLimiter(cfg, store)

    mux := http.NewServeMux()
    mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        _, _ = w.Write([]byte(`{"status":"ok"}`))
    })

    handler := middleware.RateLimitMiddleware(rl)(mux)

    srv := &http.Server{
        Addr:              fmt.Sprintf(":%d", cfg.ServerPort),
        Handler:           handler,
        ReadHeaderTimeout: 5 * time.Second,
    }

    go func() {
        log.Printf("server listening on :%d", cfg.ServerPort)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %v", err)
        }
    }()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
    <-stop

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = srv.Shutdown(ctx)
}
