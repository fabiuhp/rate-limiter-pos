package config

import (
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"
)

type RedisConfig struct {
    Addr     string
    Password string
    DB       int
}

type TokenRule struct {
    LimitPerSecond int
    BlockFor       time.Duration
}

type Config struct {
    ServerPort int
    Strategy string
    APIKeyHeader string
    IPLimitPerSecond int
    IPBlockFor       time.Duration
    TokenDefaultLimitPerSecond int
    TokenDefaultBlockFor       time.Duration
    TokenRules map[string]TokenRule
    StoreDriver string
    Redis       RedisConfig
}

func Load() Config {
    if _, err := os.Stat(".env"); err == nil {
        _ = loadDotEnv(".env")
    }

    cfg := Config{
        ServerPort:                  intFromEnv("SERVER_PORT", 8080),
        Strategy:                    strFromEnv("RATE_LIMIT_STRATEGY", "both"),
        APIKeyHeader:                strFromEnv("API_KEY_HEADER", "API_KEY"),
        IPLimitPerSecond:            intFromEnv("RATE_LIMIT_PER_SECOND", 5),
        IPBlockFor:                  durFromEnv("RATE_LIMIT_BLOCK_DURATION", mustParseDuration("5m")),
        TokenDefaultLimitPerSecond:  intFromEnv("TOKEN_DEFAULT_LIMIT_PER_SECOND", 0),
        TokenDefaultBlockFor:        durFromEnv("TOKEN_DEFAULT_BLOCK_DURATION", 0),
        TokenRules:                  parseTokenRules(strFromEnv("TOKEN_LIMITS", "")),
        StoreDriver:                 strFromEnv("STORE_DRIVER", "redis"),
        Redis: RedisConfig{
            Addr:     strFromEnv("REDIS_ADDR", "redis:6379"),
            Password: strFromEnv("REDIS_PASSWORD", ""),
            DB:       intFromEnv("REDIS_DB", 0),
        },
    }
    return cfg
}

func parseTokenRules(s string) map[string]TokenRule {
    res := map[string]TokenRule{}
    if strings.TrimSpace(s) == "" {
        return res
    }
    parts := splitCSV(s)
    for _, p := range parts {
        fields := strings.Split(p, ":")
        if len(fields) != 3 {
            continue
        }
        token := strings.TrimSpace(fields[0])
        limit, err1 := strconv.Atoi(strings.TrimSpace(fields[1]))
        blk := strings.TrimSpace(fields[2])
        block, err2 := time.ParseDuration(blk)
        if err1 != nil || err2 != nil || token == "" {
            continue
        }
        res[token] = TokenRule{LimitPerSecond: limit, BlockFor: block}
    }
    return res
}

func splitCSV(s string) []string {
    raw := strings.Split(s, ",")
    out := make([]string, 0, len(raw))
    for _, r := range raw {
        r = strings.TrimSpace(r)
        if r != "" {
            out = append(out, r)
        }
    }
    return out
}

func mustParseDuration(s string) time.Duration {
    d, err := time.ParseDuration(s)
    if err != nil {
        panic(fmt.Sprintf("invalid duration %q: %v", s, err))
    }
    return d
}

func strFromEnv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

func intFromEnv(key string, def int) int {
    if v := os.Getenv(key); v != "" {
        if n, err := strconv.Atoi(v); err == nil {
            return n
        }
    }
    return def
}

func durFromEnv(key string, def time.Duration) time.Duration {
    if v := os.Getenv(key); v != "" {
        if d, err := time.ParseDuration(v); err == nil {
            return d
        }
    }
    return def
}
func loadDotEnv(path string) error {
    b, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    lines := strings.Split(string(b), "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        kv := strings.SplitN(line, "=", 2)
        if len(kv) != 2 {
            continue
        }
        k := strings.TrimSpace(kv[0])
        v := strings.TrimSpace(kv[1])
        v = strings.Trim(v, "\"'")
        _ = os.Setenv(k, v)
    }
    return nil
}
