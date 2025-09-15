package storage

import (
    "context"
    "fmt"
    "time"

    redis "github.com/redis/go-redis/v9"
)

type RedisStore struct {
    rdb *redis.Client
}

type RedisConfig struct {
    Addr     string
    Password string
    DB       int
}

func NewRedisStore(cfg RedisConfig) (*RedisStore, error) {
    rdb := redis.NewClient(&redis.Options{Addr: cfg.Addr, Password: cfg.Password, DB: cfg.DB})
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    if err := rdb.Ping(ctx).Err(); err != nil {
        return nil, err
    }
    return &RedisStore{rdb: rdb}, nil
}

var lua = redis.NewScript(`
local count_key = KEYS[1]
local ban_key = KEYS[2]
local limit = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local block_ms = tonumber(ARGV[3])

local ttl = redis.call('PTTL', ban_key)
if ttl and ttl > 0 then
  return {0, ttl, 0, 0}
end

local current = redis.call('INCR', count_key)
if current == 1 then
  redis.call('PEXPIRE', count_key, window_ms)
end
if current > limit then
  redis.call('PEXPIRE', ban_key, block_ms)
  local bat = redis.call('PTTL', ban_key)
  return {0, bat, 0, 0}
end
local rem = limit - current
local wt = redis.call('PTTL', count_key)
local now = redis.call('TIME')
local reset_epoch_ms = (now[1] * 1000) + math.floor(now[2] / 1000) + wt
return {1, 0, rem, reset_epoch_ms}
`)

func (s *RedisStore) Attempt(scope, key string, limit int, window time.Duration, blockFor time.Duration) (AttemptResult, error) {
    ctx := context.Background()
    k := fmt.Sprintf("rl:count:%s:%s", scope, key)
    b := fmt.Sprintf("rl:block:%s:%s", scope, key)
    res, err := lua.Run(ctx, s.rdb, []string{k, b}, limit, window.Milliseconds(), blockFor.Milliseconds()).Result()
    if err != nil {
        return AttemptResult{}, err
    }
    arr, ok := res.([]interface{})
    if !ok || len(arr) != 4 {
        return AttemptResult{}, fmt.Errorf("unexpected script result: %#v", res)
    }
    allowed := arr[0].(int64) == 1
    retryAfterMs := arr[1].(int64)
    remaining := int(arr[2].(int64))
    resetEpochMs := arr[3].(int64)

    out := AttemptResult{Allowed: allowed, Remaining: remaining}
    if retryAfterMs > 0 {
        out.RetryAfter = time.Duration(retryAfterMs) * time.Millisecond
    }
    if resetEpochMs > 0 {
        out.WindowReset = time.UnixMilli(resetEpochMs)
    }
    return out, nil
}

func (s *RedisStore) Close() error { return s.rdb.Close() }
