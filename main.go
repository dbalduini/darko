package main

import (
	"context"
	"fmt"
	"github.com/dbalduini/darko/http"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/dbalduini/darko/clients"
	"github.com/dbalduini/darko/dotenv"
	"github.com/dbalduini/darko/fifo"
	"github.com/dbalduini/darko/ratelimit"
)

func main() {
	// loads dotenv file if any
	if err := dotenv.LoadFile(".env"); err != nil {
		log.Fatalln(err)
	}

	// HA
	var (
		isHA     = dotenv.GetBool("DARKO_HA_MODE")
		isMaster = dotenv.GetBool("DARKO_HA_MASTER_NODE")
		port     = dotenv.GetOrElse("DARKO_HA_MASTER_PORT", "80")
	)

	// Redis
	var (
		redisAddr    = dotenv.GetOrElse("REDIS_ADDRESS", "localhost:6379")
		redisPass    = dotenv.GetOrElse("REDIS_CON_PASSWORD", "")
		redisPool    = dotenv.GetInt("REDIS_CON_POOL")
		redisTimeout = dotenv.GetInt("REDIS_BLOCK_TIMEOUT_SECONDS")
	)

	// Rate Limiter
	var (
		points  = dotenv.GetInt("RATE_LIMIT")
		workers = dotenv.GetInt("TOTAL_WORKERS")
	)

	redisClient := clients.NewRedisClient(redisAddr, redisPass, redisPool)
	defer redisClient.Close()
	log.Println("(main) connected to redis")

	// create the queue consumer
	queue := fifo.NewRedisQueue(redisClient, time.Duration(redisTimeout)*time.Second)

	// create the context
	ctx, cancel := context.WithCancel(context.Background())

	// handle gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("(main) gracefully shutting down...")
		cancel()
	}()

	if isHA { // ha mode
		if isMaster { // master node
			log.Println("(ha:master) starting the http server...")
			startHTTPServer(ctx, queue, port)
		} else { // slave node
			pk, _ := queue.AddShardNode()
			log.Printf("(ha:follower) (pk:%d) listening", pk)
			startRateLimiter(ctx, pk, queue, points, workers)
		}
	} else { // standalone mode
		log.Printf("(standalone) listening")
		startRateLimiter(ctx, 0, queue, points, workers)
	}
}

func startRateLimiter(ctx context.Context, pk int, queue fifo.Queue, points int, workers int) {
	// start the rate limiter for this node
	dispatcher := ratelimit.NewDispatcher(ctx, points, workers)
	defer dispatcher.Shutdown()

	// start listening to the queue for new jobs
	dequeueAndDispatch(ctx, pk, queue, dispatcher)

	log.Println("(main) darko has shutdown successfully. bye!")
}

func dequeueAndDispatch(ctx context.Context, pk int, q fifo.Queue, d *ratelimit.Dispatcher) error {
	var wg sync.WaitGroup
	var err error

	topic := fmt.Sprintf("%s:%d", fifo.TopicJobsNew, pk)

	wg.Add(1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("(main) main loop has stopped")
				wg.Done()
				return
			default:
				job, err := q.Pop(topic)
				job.PartitionKey = pk
				if err == fifo.ErrEmptyQueue {
					// log.Printf("(main) %s", err)
				} else if err != nil {
					wg.Done()
					return
				} else {
					if err != nil {
						log.Printf("(main) (corr=%s) (reason=%s)", job.CorrelationID, err)
						return
					}
					d.Dispatch(job)
				}
			}
		}
	}()
	wg.Wait()

	return err
}

func startHTTPServer(ctx context.Context, queue fifo.Queue, port string) {
	var wg sync.WaitGroup

	wg.Add(1)
	srv := http.StartServer(&wg, port, queue)

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.TODO())
	}()

	log.Printf("(ha:master) server listening on port %s\n", port)
	wg.Wait()
}
