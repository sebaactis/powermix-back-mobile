package middlewares

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/sebaactis/powermix-back-mobile/internal/utils"
)

type bucket struct {
	count      int
	windowFrom time.Time
}

type RateLimiter struct {
	mu        sync.Mutex
	perWindow int
	window    time.Duration
	buckets   map[string]*bucket
}

func NewRateLimiter(perWindow int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		perWindow: perWindow,
		window:    window,
		buckets:   make(map[string]*bucket),
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)

	if err != nil {
		return r.RemoteAddr
	}

	return host
}

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)

			rl.mu.Lock()
			b, ok := rl.buckets[ip]
			now := time.Now()
			if !ok || now.Sub(b.windowFrom) >= rl.window {
				b = &bucket{count: 0, windowFrom: now}
				rl.buckets[ip] = b
			}
			if b.count >= rl.perWindow {
				rl.mu.Unlock()
				utils.WriteError(w, http.StatusTooManyRequests, "Rate Limit excedido", nil)
				return
			}
			b.count++
			rl.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}