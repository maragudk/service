package jobs_test

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/maragudk/service/jobs"
	"github.com/maragudk/service/model"
	"github.com/maragudk/service/sqltest"
)

func TestRunner_Start(t *testing.T) {
	t.Run("starts the runner and runs jobs until the context is cancelled", func(t *testing.T) {
		log, logs := newLogger()
		db := sqltest.CreateDatabase(t)

		runner := jobs.NewRunner(jobs.NewRunnerOptions{
			Log:          log,
			PollInterval: time.Millisecond,
			Queue:        db,
		})

		ctx, cancel := context.WithCancel(context.Background())

		runner.Register("test", func(ctx context.Context, m model.Map) error {
			foo, ok := m["foo"]
			require.True(t, ok)
			require.Equal(t, "bar", foo)

			cancel()
			return nil
		})

		err := db.CreateJob(context.Background(), "test", model.Map{"foo": "bar"}, time.Second)
		require.NoError(t, err)

		// This blocks until the context is cancelled by the job function
		runner.Start(ctx)

		require.Equal(t, "Starting\nRegistered jobs: [health test]\nSuccessfully ran job\n"+
			"Stopping\nStopped\n", logs.String())
	})

	t.Run("emits job metrics", func(t *testing.T) {
		db := sqltest.CreateDatabase(t)

		registry := prometheus.NewRegistry()

		runner := jobs.NewRunner(jobs.NewRunnerOptions{
			Metrics:      registry,
			PollInterval: time.Millisecond,
			Queue:        db,
		})

		ctx, cancel := context.WithCancel(context.Background())

		runner.Register("test", func(ctx context.Context, m model.Map) error {
			cancel()
			return nil
		})

		err := db.CreateJob(context.Background(), "test", model.Map{}, time.Second)
		require.NoError(t, err)

		runner.Start(ctx)

		metrics, err := registry.Gather()
		require.NoError(t, err)
		require.Len(t, metrics, 3)

		metric := metrics[0]
		require.Equal(t, "app_job_duration_seconds_total", metric.GetName())
		require.Equal(t, "name", metric.Metric[0].Label[0].GetName())
		require.Equal(t, "test", metric.Metric[0].Label[0].GetValue())
		require.Equal(t, "success", metric.Metric[0].Label[1].GetName())
		require.Equal(t, "true", metric.Metric[0].Label[1].GetValue())
		require.True(t, metric.Metric[0].Counter.GetValue() > 0)

		metric = metrics[1]
		require.Equal(t, "app_job_runner_receives_total", metric.GetName())
		require.Equal(t, "success", metric.Metric[0].Label[0].GetName())
		require.Equal(t, "true", metric.Metric[0].Label[0].GetValue())
		require.True(t, metric.Metric[0].Counter.GetValue() > 0)

		metric = metrics[2]
		require.Equal(t, "app_jobs_total", metric.GetName())
		require.Equal(t, "name", metric.Metric[0].Label[0].GetName())
		require.Equal(t, "test", metric.Metric[0].Label[0].GetValue())
		require.Equal(t, "success", metric.Metric[0].Label[1].GetName())
		require.Equal(t, "true", metric.Metric[0].Label[1].GetValue())
		require.Equal(t, float64(1), metric.Metric[0].Counter.GetValue())
	})
}

type queueMock struct {
}

func (q *queueMock) DeleteJob(ctx context.Context, id int) error {
	panic("implement me")
}

func (q *queueMock) GetJob(ctx context.Context) (*model.Job, error) {
	panic("implement me")
}

func TestRunner_Register(t *testing.T) {
	t.Run("panics on double job registration", func(t *testing.T) {
		runner := jobs.NewRunner(jobs.NewRunnerOptions{
			Queue: &queueMock{},
		})

		var panicked bool
		defer func() {
			if rec := recover(); rec != nil {
				panicked = true
			}
			require.True(t, panicked)
		}()

		runner.Register("foo", func(ctx context.Context, message model.Map) error {
			return nil
		})

		runner.Register("foo", func(ctx context.Context, message model.Map) error {
			return nil
		})
	})
}

func newLogger() (*log.Logger, *strings.Builder) {
	var s strings.Builder
	return log.New(&s, "", 0), &s
}
