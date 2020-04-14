package ratelimit

import (
	"context"
	"errors"
	"github.com/dbalduini/darko/fifo"
	"log"
	"sync"
	"time"

	"github.com/dbalduini/darko/http"
)

const (
	// MaxWorkersCount the max number of workers allowed to run
	MaxWorkersCount = 99
)

var (
	ErrMaxWorkers = errors.New("max number of workers allowed is 99")

	template = "(shard=%02d)(worker=%02d)(token=%d)(job=%s)(corr=%s)(status=%d)(time=%.3fs)"
)

// Work is a channel of Jobs
type Work chan fifo.Job

// Dispatcher uses round robin to dispatch messages
type Dispatcher struct {
	bucket    *TokenBucket
	workers   []Work
	nrWorkers int
	wg        sync.WaitGroup
}

func NewDispatcher(nrWorkers int, bucket *TokenBucket) *Dispatcher {
	if nrWorkers > MaxWorkersCount {
		log.Fatalln("max workers allowed is 99")
	}
	return &Dispatcher{
		bucket:    bucket,
		workers:   make([]Work, nrWorkers, nrWorkers),
		nrWorkers: nrWorkers,
	}
}

// Dispatch the payload to the correct partition worker.
func (d *Dispatcher) Dispatch(job fifo.Job) {
	i := selectWorkerIndex(job, d.nrWorkers)
	d.workers[i] <- job
}

// selectWorkerIndex returns the worker id to process the current Job based on the hash of the Job
func selectWorkerIndex(j fifo.Job, n int) int {
	return j.Hash() % n
}

func (d *Dispatcher) SpawnWorkers(ctx context.Context) {
	d.wg = sync.WaitGroup{}
	for i := 0; i < d.nrWorkers; i++ {
		d.wg.Add(1)
		d.workers[i] = make(Work, 1)
		w := newWorker(i, d.bucket, d.workers[i])
		w.run(&d.wg)
	}
}

func (d *Dispatcher) Shutdown() {
	log.Println("waiting for workers to stop")
	for i := 0; i < d.nrWorkers; i++ {
		close(d.workers[i])
	}
	d.wg.Wait()
	log.Println("all workers have stopped")
}

type worker struct {
	id     int
	bucket *TokenBucket
	work   Work
}

func newWorker(id int, bucket *TokenBucket, work Work) *worker {
	return &worker{id, bucket, work}
}

func (w *worker) run(wg *sync.WaitGroup) {
	log.Printf("(worker=%02d) new worker started", w.id)
	go func() {
		for p := range w.work {
			w.deliverWithRateLimiter(p)
		}
		log.Printf("(worker=%02d) worker has stopped", w.id)
		wg.Done()
	}()
}

func (w *worker) deliverWithRateLimiter(job fifo.Job) {
	tk := w.bucket.Take()

	start := time.Now()
	status, err := http.PostCallback(job.CallbackURL, job.NewPayloadReader())
	if err != nil {
		log.Println("PostCallback error", err)
	}
	end := time.Since(start).Seconds()

	log.Printf(template, job.PartitionKey, w.id, tk, job.ID, job.CorrelationID, status, end)
}
