package ratelimit

import (
	"context"
	"time"
)

// TokenBucket holds data for the token-bucket rate limiter algorithm.
type TokenBucket struct {
	tokens chan int
	points int
}

// NewTokenBucket returns a new TokenBucket pointer
func NewTokenBucket(points int) *TokenBucket {
	tokens := make(chan int, points)
	return &TokenBucket{tokens, points}
}

// Take a token from the bucket and blocks if the bucket is empty.
func (b *TokenBucket) Take() int {
	tk := <-b.tokens
	return tk
}

// Drain empty the token bucket
func (b *TokenBucket) Drain() {
	for {
		select {
		case <-b.tokens:
		default:
			return
		}
	}
}

// fill the bucket with tokens
func (b *TokenBucket) Fill() {
	for i := 0; i < b.points; i++ {
		b.tokens <- i
	}
}

// StartTicker starts a token bucket refresher that runs every d duration.
func (b *TokenBucket) StartRefresher(ctx context.Context, d time.Duration) {
	var t time.Time

	// best effort to get closer to the first millisecond of the next second
	t = time.Now().Truncate(time.Second)
	t = t.Add(time.Second)
	time.Sleep(time.Until(t))

	// start a ticker for every second
	ticker := time.NewTicker(d)

	// refresh token bucket
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				close(b.tokens)
				return
			case <-ticker.C:
				// log.Printf("(DEBUG) refresh token bucket at %s", t.Truncate(time.Nanosecond))
				b.Drain()
				b.Fill()
			}
		}
	}()
}
