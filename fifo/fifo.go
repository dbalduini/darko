package fifo

import (
	"errors"
)

var (
	// ErrEmptyQueue is returned when the queue is empty
	ErrEmptyQueue = errors.New("the queue is empty or does not exists")

	TopicJobsNew       = "darko:jobs:new"
	TopicJobsProcessed = "darko:jobs:processed"
	TopicJobsFailed    = "darko:jobs:failed"

)

// Queue defines a FIFO queue interface
type Queue interface {
	// Pop dequeue the next item of the topic.
	Pop(topic string) (string, error)
	// Push enqueues the item,
	Push(topic string, pack string) error
}
