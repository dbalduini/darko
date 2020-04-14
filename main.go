package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/dbalduini/darko/clients"
	"github.com/dbalduini/darko/dotenv"
	"github.com/dbalduini/darko/fifo"
	"github.com/dbalduini/darko/http"
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
	)

	// Redis
	var (
		redisAddr    = dotenv.GetOrElse("REDIS_ADDRESS", "localhost:6379").String()
		redisPass    = dotenv.GetOrElse("REDIS_CON_PASSWORD", "").String()
		redisPool    = dotenv.GetOrElse("REDIS_CON_POOL", "10").Int()
		redisTimeout = dotenv.GetOrElse("REDIS_BLOCK_TIMEOUT_SECONDS", "1").Int()
	)

	redisClient := clients.NewRedisClient(redisAddr, redisPass, redisPool)
	defer redisClient.Close()
	log.Println("(main) connected to redis")

	// create redis queue
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

	var wg sync.WaitGroup

	if isHA { // ha mode
		if isMaster { // master node
			startHTTPServer(&wg, ctx, queue)
		} else { // slave node
			startRateLimiter(&wg, ctx, queue)
		}
	} else { // standalone mode
		startRateLimiter(&wg, ctx, queue)
		startHTTPServer(&wg, ctx, queue)
	}

	wg.Wait()
	log.Println("(main) darko has shutdown successfully. bye!")
}

func startRateLimiter(wg *sync.WaitGroup, ctx context.Context, queue fifo.Queue) {
	wg.Add(1)

	var (
		points  = dotenv.MustGet("DARKO_RATE_LIMIT_POINTS").Int()
		workers = dotenv.MustGet("DARKO_RATE_LIMIT_WORKERS_COUNT").Int()
		shard   = dotenv.GetOrElse("DARKO_RATE_LIMIT_SHARD", "0").Int()
		topic   = fmt.Sprintf("%s:%d", fifo.TopicJobsNew, shard)
	)

	// start the rate limiter for this node
	rateLimiter := ratelimit.NewRateLimiter(shard, points, workers)

	go func() {
		defer wg.Done()
		if err := rateLimiter.Start(ctx, queue, topic); err != nil {
			log.Println("(ha:master) error listen queue", err)
		}
	}()
}

func startHTTPServer(wg *sync.WaitGroup, ctx context.Context, queue fifo.Queue) {
	wg.Add(1)

	var (
		port   = dotenv.GetOrElse("DARKO_SERVER_PORT", "8888").String()
		shards = dotenv.GetOrElse("DARKO_SERVER_SHARDS_COUNT", "1").Int()
	)

	srv := http.StartServer(wg, port, queue, shards)

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.TODO())
	}()

	log.Printf("(main) server listening on port %s\n", port)
}
