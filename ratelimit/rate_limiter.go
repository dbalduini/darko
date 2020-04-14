package ratelimit

import (
	"context"
	"github.com/dbalduini/darko/fifo"
	"log"
	"time"
)

type RateLimiter struct {
	Shard     int
	Points    int
	NrWorkers int
}

func NewRateLimiter(shard int, points int, nrWorkers int) *RateLimiter {
	rl := RateLimiter{shard, points, nrWorkers}
	return &rl
}

// Start listens to the queue and dispatch all new incoming jobs.
func (rl *RateLimiter) Start(ctx context.Context, queue fifo.Queue, topic string) (err error) {
	var (
		data string
		job  fifo.Job
	)

	// create a new filled token bucket
	bucket := NewTokenBucket(rl.Points)
	bucket.Fill()
	bucket.StartRefresher(ctx, time.Second)

	// create dispatcher and spawn workers
	dispatcher := NewDispatcher(rl.NrWorkers, bucket)
	dispatcher.SpawnWorkers(ctx)

	// run rate limiter loop
loop:
	for {
		select {
		case <-ctx.Done():
			err = nil
			log.Println("(rate_limiter) main loop has stopped")
			dispatcher.Shutdown()
			log.Println("(rate_limiter) dispatcher was shutdown")
			break loop
		default:
			data, err = queue.Pop(topic)

			if err == fifo.ErrEmptyQueue {
				continue
			} else if err != nil {
				break loop
			}

			err = fifo.Unpack(data, &job)
			if err != nil {
				log.Println("error unpacking job", err)
				continue
			}

			log.Printf("(shard=%d)(pk=%d)(job=%s) job dispatched\n", rl.Shard, job.PartitionKey, job.ID)
			dispatcher.Dispatch(job)
		}
	}

	return
}
