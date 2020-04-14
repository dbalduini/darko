package fifo

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"hash/fnv"
	"io"
	"strings"
)

// Job holds data for creating jobs.
// A job defines something that will be queued and dequeued. In HA mode, a job
// will also be uniform partitioned on shards.
type Job struct {
	// ID is the job unique id. Set by the server.
	ID string
	// PartitionKey is the partition id this job belongs to. Set by the server.
	PartitionKey int
	// PrimaryKey is the primary key.
	// It is used to guarantee the order of the requests and balance the work load.
	PrimaryKey string `json:"primary_key"`
	// CallbackURL is the service endpoint to send the payload back to.
	CallbackURL string `json:"callback_url"`
	// CorrelationID is a unique identifier attached to the request by the client that allow
	// reference to a particular transaction or event chain.
	CorrelationID string `json:"correlation_id"`
	// Payload with the content encoded as base64 that will be forwarded on the client on the callback.
	Payload string `json:"payload"`
}

// Hash returns hash code for the primary key
func (j Job) Hash() int {
	h := fnv.New32a()
	h.Write([]byte(j.PrimaryKey))
	return int(h.Sum32())
}

func (j Job) NewPayloadReader() io.Reader {
	return strings.NewReader(j.Payload)
}

// Unpack the job from the queue after dequeue.
func Unpack(pack string, job *Job) error {
	buf, err := base64.StdEncoding.DecodeString(pack)
	if err != nil {
		return err
	}
	d := gob.NewDecoder(bytes.NewBuffer(buf))
	return d.Decode(job)
}

// Pack the job as base64 string to be enqueued
func Pack(job Job) (string, error) {
	var buf bytes.Buffer
	e := gob.NewEncoder(&buf)
	if err := e.Encode(job); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
