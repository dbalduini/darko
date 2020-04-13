package fifo

import (
	"errors"
	"github.com/dbalduini/darko/shard"
)

var (
	// ErrEmptyQueue is returned when the queue is empty
	ErrEmptyQueue = errors.New("the queue is empty or does not exists")

	TopicJobsNew       = "darko:jobs:new"
	TopicJobsProcessed = "darko:jobs:processed"
	TopicJobsFailed    = "darko:jobs:failed"
	TopicShardsNode    = "darko:shards:node"
)

// Queue defines a FIFO queue interface
type Queue interface {
	// Pop dequeue the next item of the topic.
	Pop(topic string) (shard.Job, error)
	// Push enqueues the item,
	Push(topic string, p shard.Job) error

	AddShardNode() (int, error)

	ShardCount() (int, error)
}
