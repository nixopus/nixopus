package redisclient

import (
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// New creates a new Redis v8 client from a redis URL (e.g. redis://localhost:6379).
func New(redisURL string) (*redis.Client, error) {
	log.Printf("Attempting to connect to Redis at %s", redisURL)
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("Failed to parse Redis URL: %v", err)
		return nil, err
	}
	// Sensible defaults; callers can adjust via WithOptions later if needed.
	opt.MinIdleConns = 10
	opt.DialTimeout = 5 * time.Second
	opt.ReadTimeout = 3 * time.Second
	opt.WriteTimeout = 3 * time.Second
	return redis.NewClient(opt), nil
}
