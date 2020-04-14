package http

import (
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"log"
	"net/http"
	"sync"

	"github.com/dbalduini/darko/fifo"
)

// StartServer starts a http server to create new jobs.
// On HA mode, the master will up the http server to receive new jobs.
// On standalone mode, the server and the queue listener will run on the same process.
func StartServer(wg *sync.WaitGroup, port string, queue fifo.Queue, shards int) *http.Server {
	srv := &http.Server{Addr: ":" + port}

	http.Handle("/jobs", &jobHandler{queue, shards})

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
	queue  fifo.Queue
	shards int
}

func (h *jobHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var job fifo.Job
	var id string

	if req.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := json.NewDecoder(req.Body).Decode(&job); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if job.PrimaryKey == "" {
		http.Error(w, "primary_key cannot be empty", http.StatusBadRequest)
		return
	}

	// get the correct partition to send the job
	pk := job.Hash() % h.shards

	// set job info
	job.ID = uuid.NewV4().String()
	job.PartitionKey = pk

	data, _ := fifo.Pack(job)

	// put the job on the queue
	topic := fmt.Sprintf("%s:%d", fifo.TopicJobsNew, pk)
	if err := h.queue.Push(topic, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(fmt.Sprintf("{\"id\":\"%s\"}", id)))
}
