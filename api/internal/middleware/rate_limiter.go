package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/types"
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

// RateLimiter is a middleware that prevents excessive requests from the same
// IP address in a given time frame. The rate limiter allows 2 requests every
// 5 seconds. If the rate limit is exceeded, a 429 Too Many Requests response
// is returned. The middleware also cleans up inactive clients every minute,
// removing them from the memory after 3 minutes of inactivity.
func RateLimiter(next http.Handler) http.Handler {
	if len(clients) == 0 {
		StartCleanupTask()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
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

		fmt.Printf("Request from: %s (raw: %s), Path: %s\n", ip, r.RemoteAddr, r.URL.Path)

		clientsMtx.Lock()

		c, exists := clients[ip]
		if !exists {
			c = &client{
				limiter:  rate.NewLimiter(2, 5),
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
			message := types.Response{
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
