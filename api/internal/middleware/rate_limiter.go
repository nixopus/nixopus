package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var (
	clients    = make(map[string]*client)
	clientsMtx sync.Mutex
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimitResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func StartCleanupTask() {
	go func() {
		for {
			time.Sleep(time.Minute)
			cleanupClients()
		}
	}()
}

func cleanupClients() {
	clientsMtx.Lock()
	defer clientsMtx.Unlock()

	for ip, client := range clients {
		if time.Since(client.lastSeen) > 3*time.Minute {
			fmt.Printf("Removing inactive client: %s\n", ip)
			delete(clients, ip)
		}
	}
}

func extractIP(remoteAddr string) string {
	ip := remoteAddr
	if strings.Contains(ip, ":") {
		ipParts := strings.Split(ip, ":")
		if len(ipParts) > 0 {
			if strings.Contains(ip, "[") {
				endBracket := strings.LastIndex(ip, "]")
				if endBracket > 0 {
					ip = ip[:endBracket+1]
				}
			} else {
				ip = strings.Join(ipParts[:len(ipParts)-1], ":")
			}
		}
	}
	return ip
}

// RateLimiter is a middleware that prevents excessive requests from the same
// IP address in a given time frame. The rate limiter allows 5 requests per
// second with a burst of 10. If the rate limit is exceeded, a 429 Too Many
// Requests response is returned.
func RateLimiter(next http.Handler) http.Handler {
	if len(clients) == 0 {
		StartCleanupTask()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r.RemoteAddr)

		fmt.Printf("Request from: %s (raw: %s), Path: %s\n", ip, r.RemoteAddr, r.URL.Path)

		clientsMtx.Lock()

		c, exists := clients[ip]
		if !exists {
			c = &client{
				limiter:  rate.NewLimiter(5, 10),
				lastSeen: time.Now(),
			}
			clients[ip] = c
		} else {
			c.lastSeen = time.Now()
		}

		allowed := c.limiter.Allow()

		if !allowed {
			clientsMtx.Unlock()
			fmt.Printf("Rate limit exceeded for IP: %s\n", ip)
			message := rateLimitResponse{
				Status:  "Request Failed",
				Message: "The API is at capacity, try again later.",
			}
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(&message)
			return
		}

		clientsMtx.Unlock()
		next.ServeHTTP(w, r)
	})
}

// NewRateLimiterWithConfig creates a rate limiter middleware with custom
// rate (requests per second) and burst size. Each IP gets its own bucket.
// Use tight values for public endpoints vulnerable to spam.
func NewRateLimiterWithConfig(rps float64, burst int) func(http.Handler) http.Handler {
	var (
		rlClients = make(map[string]*client)
		rlMtx     sync.Mutex
		once      sync.Once
	)

	startCleanup := func() {
		go func() {
			for {
				time.Sleep(time.Minute)
				rlMtx.Lock()
				for ip, c := range rlClients {
					if time.Since(c.lastSeen) > 5*time.Minute {
						delete(rlClients, ip)
					}
				}
				rlMtx.Unlock()
			}
		}()
	}

	return func(next http.Handler) http.Handler {
		once.Do(startCleanup)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r.RemoteAddr)

			rlMtx.Lock()

			c, exists := rlClients[ip]
			if !exists {
				c = &client{
					limiter:  rate.NewLimiter(rate.Limit(rps), burst),
					lastSeen: time.Now(),
				}
				rlClients[ip] = c
			} else {
				c.lastSeen = time.Now()
			}

			allowed := c.limiter.Allow()

			if !allowed {
				rlMtx.Unlock()
				msg := rateLimitResponse{
					Status:  "Request Failed",
					Message: "Too many requests, try again later.",
				}
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(&msg)
				return
			}

			rlMtx.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}
