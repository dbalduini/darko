package shard

import (
	"fmt"
	"hash/fnv"
	"strings"
)

// Job holds data for creating jobs.
// A job defines something that will be queued and dequeued. In HA mode, a job
// will also be uniform partitioned on shards.
type Job struct {
	// PK is the primary key.
	// It is used to guarantee the order of the requests and balance the work load.
	PK string `json:"primary_key"`
	// CorrelationID is a unique identifier attached to the request by the client that allow
	// reference to a particular transaction or event chain.
	CorrelationID string `json:"correlation_id"`
	// Payload with the content encoded as base64 that will be forwarded on the client on the callback.
	Payload string `json:"payload"`

	PartitionKey int
}

// Hash returns hash code for the primary key
func (j Job) Hash() int {
	h := fnv.New32a()
	h.Write([]byte(j.PK))
	return int(h.Sum32())
}

// Unpack the job from the queue after dequeue.
// PK:CorrelationID:base64(content)
func Unpack(str string, job *Job) error {
	// split the payload
	parts := strings.SplitN(str, ":", 3)
	if len(parts) != 3 {
		return fmt.Errorf("could not split the payload into 3 parts")
	}
	job.PK = parts[0]
	job.CorrelationID = parts[1]
	job.Payload = parts[2]
	return nil
}

// Pack the job as string to be enqueued
func Pack(job Job) string {
	template := "%s:%s:%s"
	return fmt.Sprintf(template, job.PK, job.CorrelationID, job.CorrelationID)
}

// GetPartitionKey returns the job partition key based on hash(job pk)
func GetPartitionKey(job Job, shards int) string {
	p := job.Hash() % shards
	return fmt.Sprintf("%d", p)
}
