// Package jobs has a Runner that can run registered jobs in parallel.
package jobs

import (
	"context"
	"io"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/maragudk/service/email"
	"github.com/maragudk/service/model"
	"github.com/maragudk/service/sql"
)

// Runner runs jobs.
type Runner struct {
	currentJobCount     int
	currentJobCountLock sync.RWMutex
	database            *sql.Database
	emailSender         *email.Sender
	jobCount            *prometheus.CounterVec
	jobDuration         *prometheus.CounterVec
	jobCountLimit       int
	jobs                map[string]Func
	log                 *log.Logger
	pollInterval        time.Duration
	queue               queue
	runnerReceives      *prometheus.CounterVec
}

type NewRunnerOptions struct {
	Database     *sql.Database
	EmailSender  *email.Sender
	JobLimit     int
	Log          *log.Logger
	Metrics      *prometheus.Registry
	PollInterval time.Duration
	Queue        queue
}

type queue interface {
	DeleteJob(ctx context.Context, id int) error
	GetJob(ctx context.Context) (*model.Job, error)
}

func NewRunner(opts NewRunnerOptions) *Runner {
	if opts.Log == nil {
		opts.Log = log.New(io.Discard, "", 0)
	}

	if opts.Metrics == nil {
		opts.Metrics = prometheus.NewRegistry()
	}

	if opts.JobLimit == 0 {
		opts.JobLimit = 1
	}

	if opts.PollInterval == 0 {
		opts.PollInterval = time.Second
	}

	jobCount := promauto.With(opts.Metrics).NewCounterVec(prometheus.CounterOpts{
		Name: "app_jobs_total",
	}, []string{"name", "success"})

	jobDuration := promauto.With(opts.Metrics).NewCounterVec(prometheus.CounterOpts{
		Name: "app_job_duration_seconds_total",
	}, []string{"name", "success"})

	runnerReceives := promauto.With(opts.Metrics).NewCounterVec(prometheus.CounterOpts{
		Name: "app_job_runner_receives_total",
	}, []string{"success"})

	return &Runner{
		database:       opts.Database,
		emailSender:    opts.EmailSender,
		jobCount:       jobCount,
		jobDuration:    jobDuration,
		jobCountLimit:  opts.JobLimit,
		jobs:           map[string]Func{},
		log:            opts.Log,
		pollInterval:   opts.PollInterval,
		queue:          opts.Queue,
		runnerReceives: runnerReceives,
	}
}

// Func is the actual work to do in a job.
// The given context is the root context of the runner, which may be cancelled.
// It also has a timeout.
type Func = func(context.Context, model.Map) error

// Start the Runner, blocking until the given context is cancelled.
func (r *Runner) Start(ctx context.Context) {
	r.log.Println("Starting")
	r.registerJobs()

	var names []string
	for k := range r.jobs {
		names = append(names, k)
	}
	sort.Strings(names)

	r.log.Println("Registered jobs:", names)

	var wg sync.WaitGroup

	ticker := time.NewTicker(r.pollInterval)

	for {
		select {
		case <-ctx.Done():
			r.log.Println("Stopping")
			ticker.Stop()
			wg.Wait()
			r.log.Println("Stopped")
			return
		case <-ticker.C:
			r.receiveAndRun(ctx, &wg)
		}
	}
}

// receiveAndRun jobs.
func (r *Runner) receiveAndRun(ctx context.Context, wg *sync.WaitGroup) {
	r.currentJobCountLock.RLock()
	if r.currentJobCount == r.jobCountLimit {
		r.currentJobCountLock.RUnlock()
		return
	} else {
		r.currentJobCountLock.RUnlock()
	}

	j, err := r.queue.GetJob(ctx)
	if err != nil {
		r.runnerReceives.WithLabelValues("false").Inc()
		// Sleep a bit to not hammer the queue if there's an error with it
		time.Sleep(time.Second)
		return
	}

	// If there was no job there is nothing to do
	if j == nil {
		r.runnerReceives.WithLabelValues("true").Inc()
		return
	}

	job, ok := r.jobs[j.Name]
	if !ok {
		r.runnerReceives.WithLabelValues("false").Inc()
		r.log.Println("No job with this name:", j.Name)
		return
	}

	r.runnerReceives.WithLabelValues("true").Inc()

	r.currentJobCountLock.Lock()
	r.currentJobCount++
	r.currentJobCountLock.Unlock()

	wg.Add(1)
	go func() {
		defer wg.Done()

		defer func() {
			r.currentJobCountLock.Lock()
			r.currentJobCount--
			r.currentJobCountLock.Unlock()
		}()

		defer func() {
			if rec := recover(); rec != nil {
				r.jobCount.WithLabelValues(j.Name, "false").Inc()
				r.log.Println("Recovered from panic in job:", rec)
			}
		}()

		jobCtx, cancel := context.WithTimeout(ctx, j.Timeout)
		defer cancel()

		before := time.Now()
		err := job(jobCtx, j.Payload)
		duration := time.Since(before)

		success := strconv.FormatBool(err == nil)
		r.jobCount.WithLabelValues(j.Name, success).Inc()
		r.jobDuration.WithLabelValues(j.Name, success).Add(duration.Seconds())

		if err != nil {
			r.log.Println("Error running job:", err)
			return
		}

		// We use context.Background as the parent context instead of the existing ctx, because if we've come
		// this far we don't want the deletion to be cancelled.
		deleteCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := r.queue.DeleteJob(deleteCtx, j.ID); err != nil {
			r.log.Println("Error deleting job, it will be repeated:", err)
		}
	}()
}

// registry provides a way to Register jobs by name.
type registry interface {
	Register(name string, fn Func)
}

// Register job by name. Satisfies the registry interface.
func (r *Runner) Register(name string, j Func) {
	if _, ok := r.jobs[name]; ok {
		panic("there is already a job with this name: " + name)
	}
	r.jobs[name] = j
}
