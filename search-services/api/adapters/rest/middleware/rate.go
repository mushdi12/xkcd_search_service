package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

var (
	rateLimitersMu sync.Mutex
	rateLimiters   = make(map[int]*rate.Limiter)
)

func getRateLimiter(rps int) *rate.Limiter {
	rateLimitersMu.Lock()
	defer rateLimitersMu.Unlock()

	if l, ok := rateLimiters[rps]; ok {
		return l
	}

	// burst = 1, чтобы не было большого стартового всплеска RPS
	l := rate.NewLimiter(rate.Limit(rps), 1)
	rateLimiters[rps] = l
	return l
}

func Rate(next http.HandlerFunc, rps int) http.HandlerFunc {
	if rps <= 0 {
		return next
	}

	limiter := getRateLimiter(rps)

	return func(w http.ResponseWriter, r *http.Request) {
		if err := limiter.Wait(r.Context()); err != nil {
			http.Error(w, "rate limit error", http.StatusInternalServerError)
			return
		}
		next(w, r)
	}
}
