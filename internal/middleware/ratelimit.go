package middleware

import (
    "net/http"
    "strconv"

    "github.com/fabiuhp/rate-limiter-pos/internal/limiter"
)

func RateLimitMiddleware(l *limiter.Limiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            dec, err := l.Evaluate(r)
            if err != nil {
                http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
                return
            }
            if !dec.Allowed {
                if dec.RetryAfter > 0 {
                    w.Header().Set("Retry-After", strconv.Itoa(int(dec.RetryAfter.Seconds())))
                }
                w.WriteHeader(http.StatusTooManyRequests)
                _, _ = w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
