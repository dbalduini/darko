package fifo

import (
	"time"

	"github.com/go-redis/redis"
)

// RedisQueue implements the Queue interface using Redis database as queue.
type RedisQueue struct {
	cli          *redis.Client
	blockTimeout time.Duration
}

// NewRedisQueue returns a RedisQueue.
func NewRedisQueue(cli *redis.Client, blockTimeout time.Duration) *RedisQueue {
	return &RedisQueue{cli, blockTimeout}
}

// Pop returns the left most item of the list.
// The command BLPop is used to block the queue when it is empty until the timeout.
// The item must be added by the producer with RPUSH.
func (r *RedisQueue) Pop(topic string) (string, error) {
	res, err := r.cli.BLPop(r.blockTimeout, topic).Result()
	if err == redis.Nil {
		return "", ErrEmptyQueue
	}
	if err != nil {
		return "", err
	}
	// the zero index contains the list name
	return res[1], err
}

// Push enqueues the item pushing it on the right side of the list.
// This is it can be dequeue my LPOP or BLPop in FIFO order.
func (r *RedisQueue) Push(topic string, pack string) error {
	_, err := r.cli.RPush(topic, pack).Result()
	return err
}

