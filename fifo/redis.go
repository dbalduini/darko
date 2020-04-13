package fifo

import (
	"time"

	"github.com/go-redis/redis"

	"github.com/dbalduini/darko/shard"
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
func (r *RedisQueue) Pop(topic string) (shard.Job, error) {
	job := shard.Job{}
	res, err := r.cli.BLPop(r.blockTimeout, topic).Result()
	if err == redis.Nil {
		return job, ErrEmptyQueue
	}
	if err != nil {
		return job, err
	}
	// the zero index contains the list name
	s := res[1]
	err = shard.Unpack(s, &job)
	return job, err
}

// Push enqueues the item pushing it on the right side of the list.
// This is it can be dequeue my LPOP or BLPop in FIFO order.
func (r *RedisQueue) Push(topic string, job shard.Job) error {
	s := shard.Pack(job)
	_, err := r.cli.RPush(topic, s).Result()
	return err
}

// AddShardNode register the node and returns the partition key
func (r *RedisQueue) AddShardNode() (int, error) {
	p, err := r.cli.Incr(TopicShardsNode).Result()
	return int(p), err
}

// AddShardNode register the node and returns the partition key
func (r *RedisQueue) ShardCount() (int, error) {
	p, err := r.cli.Get(TopicShardsNode).Int()
	return p, err
}
