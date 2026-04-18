package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/nabsk911/chronify/internal/utils"
	"golang.org/x/time/rate"
)

type Client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	clients = make(map[string]*Client)
	mu      sync.Mutex
)

func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := clients[ip]
	if !exists {
		limit := rate.Every(1 * time.Minute)
		limiter := rate.NewLimiter(limit, 1)

		clients[ip] = &Client{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func init() {
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

func LimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		if !getLimiter(ip).Allow() {
			utils.WriteJSON(w, http.StatusTooManyRequests, utils.Envelope{"message": "Too many requests"})
			return
		}
		next.ServeHTTP(w, r)
	}
}
