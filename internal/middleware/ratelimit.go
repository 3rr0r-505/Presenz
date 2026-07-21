// internal/middleware/ratelimit.go

package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/3rr0r-505/Presenz/internal/models"
)

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*rate.Limiter
	r        rate.Limit
	burst    int
}

// NewRateLimiter parses a "N/period" string (e.g. "3/minute") into a
// per-IP token bucket limiter, mirroring slowapi's rate_limit config.
func NewRateLimiter(spec string) (*RateLimiter, error) {
	parts := strings.SplitN(spec, "/", 2)
	n, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}

	var per time.Duration
	switch parts[1] {
	case "second":
		per = time.Second
	case "minute":
		per = time.Minute
	case "hour":
		per = time.Hour
	}

	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		r:        rate.Every(per / time.Duration(n)),
		burst:    n,
	}, nil
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	l, ok := rl.visitors[ip]
	if !ok {
		l = rate.NewLimiter(rl.r, rl.burst)
		rl.visitors[ip] = l
	}
	return l
}

// Limit wraps a single handler (not the whole mux) with per-IP rate
// limiting, matching slowapi's @limiter.limit scoping to POST /submit only.
func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}

		if !rl.getLimiter(host).Allow() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(models.ErrorResponse{Detail: "Rate limit exceeded"})
			return
		}
		next(w, r)
	}
}
