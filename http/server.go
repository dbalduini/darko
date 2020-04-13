package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/dbalduini/darko/fifo"
	"github.com/dbalduini/darko/shard"
)

// StartServer starts a http server to create new jobs.
// On HA mode, the master will up the http server to receive new jobs.
// On standalone mode, the server and the queue listener will run on the same process.
func StartServer(wg *sync.WaitGroup, port string, queue fifo.Queue) *http.Server {
	srv := &http.Server{Addr: ":" + port}

	http.Handle("/jobs", &jobHandler{queue})

	go func() {
		defer wg.Done() // let main know we are done cleaning up

		// always returns error. ErrServerClosed on graceful close
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln("(ha:master) error starting the server", err)
		}
	}()

	return srv
}

type jobHandler struct {
	queue fifo.Queue
}

func (h *jobHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var job shard.Job
	var id string

	log.Printf("%s\t%s", req.Method, req.URL.Path)

	if req.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(req.Body).Decode(&job); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if job.PK == "" {
		http.Error(w, "primary_key cannot be empty", http.StatusBadRequest)
		return
	}

	shards, _ := h.queue.ShardCount()

	p := shard.GetPartitionKey(job, shards)
	topic := fmt.Sprintf("%s:%s", fifo.TopicJobsNew, p)

	if err := h.queue.Push(topic, job); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf("{\"id\":\"%s\"}", id)))
}
