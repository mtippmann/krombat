package handlers

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

type rateLimiter struct {
	mu       sync.Mutex
	last     map[string]time.Time
	interval time.Duration
}

func newRateLimiter(interval time.Duration) *rateLimiter {
	return &rateLimiter{last: make(map[string]time.Time), interval: interval}
}

func (rl *rateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	if t, ok := rl.last[key]; ok && now.Sub(t) < rl.interval {
		return false
	}
	rl.last[key] = now
	evictBefore := now.Add(-rl.interval * 10)
	for k, t := range rl.last {
		if t.Before(evictBefore) {
			delete(rl.last, k)
		}
	}
	return true
}

func (rl *rateLimiter) Wrap(next http.HandlerFunc, keyFn func(*http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rl.Allow(keyFn(r)) {
			w.Header().Set("Retry-After", strconv.Itoa(max(1, int(rl.interval.Seconds()))))
			writeError(w, "rate limit exceeded, try again shortly", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
