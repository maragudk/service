package sql_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/maragudk/service/model"
	"github.com/maragudk/service/sqltest"
)

func TestDatabase_CreateJob(t *testing.T) {
	t.Run("can create job", func(t *testing.T) {
		db := sqltest.CreateDatabase(t)

		err := db.CreateJob(context.Background(), "test", model.Map{}, time.Minute)
		require.NoError(t, err)

		var jobs []model.Job
		err = db.DB.Select(&jobs, `select * from jobs`)
		require.NoError(t, err)

		require.Len(t, jobs, 1)
		require.Equal(t, "test", jobs[0].Name)
		require.Equal(t, model.Map{}, jobs[0].Payload)
		require.Equal(t, time.Minute, jobs[0].Timeout)
		require.WithinDuration(t, time.Now(), jobs[0].Run.T, time.Second)
		require.Nil(t, jobs[0].Received)
		require.WithinDuration(t, time.Now(), jobs[0].Created.T, time.Second)
		require.WithinDuration(t, time.Now(), jobs[0].Updated.T, time.Second)
	})
}

func TestDatabase_GetJob(t *testing.T) {
	t.Run("gets job to run now", func(t *testing.T) {
		db := sqltest.CreateDatabase(t)

		err := db.CreateJob(context.Background(), "test", model.Map{}, time.Minute)
		require.NoError(t, err)

		job, err := db.GetJob(context.Background())
		require.NoError(t, err)
		require.NotNil(t, job)

		require.Equal(t, model.Map{}, job.Payload)
		require.Equal(t, time.Minute, job.Timeout)
		require.WithinDuration(t, time.Now(), job.Run.T, time.Second)
		require.WithinDuration(t, time.Now(), job.Received.T, time.Second)
		require.WithinDuration(t, time.Now(), job.Created.T, time.Second)
		require.WithinDuration(t, time.Now(), job.Updated.T, time.Second)
	})

	t.Run("doesn't get a job twice immediately", func(t *testing.T) {
		db := sqltest.CreateDatabase(t)

		err := db.CreateJob(context.Background(), "test", model.Map{}, time.Minute)
		require.NoError(t, err)

		job, err := db.GetJob(context.Background())
		require.NoError(t, err)
		require.NotNil(t, job)

		job, err = db.GetJob(context.Background())
		require.NoError(t, err)
		require.Nil(t, job)
	})

	t.Run("returns the job again after timeout", func(t *testing.T) {
		db := sqltest.CreateDatabase(t)

		err := db.CreateJob(context.Background(), "test", model.Map{}, time.Millisecond)
		require.NoError(t, err)

		job, err := db.GetJob(context.Background())
		require.NoError(t, err)
		require.NotNil(t, job)

		time.Sleep(time.Millisecond)

		job, err = db.GetJob(context.Background())
		require.NoError(t, err)
		require.NotNil(t, job)
	})

	t.Run("doesn't get a job created for later", func(t *testing.T) {
		db := sqltest.CreateDatabase(t)

		err := db.CreateJobForLater(context.Background(), "test", model.Map{}, time.Minute, time.Minute)
		require.NoError(t, err)

		job, err := db.GetJob(context.Background())
		require.NoError(t, err)
		require.Nil(t, job)
	})
}

func TestDatabase_DeleteJob(t *testing.T) {
	t.Run("deletes a job", func(t *testing.T) {
		db := sqltest.CreateDatabase(t)

		err := db.CreateJob(context.Background(), "test", model.Map{}, 0)
		require.NoError(t, err)

		job, err := db.GetJob(context.Background())
		require.NoError(t, err)
		require.NotNil(t, job)

		err = db.DeleteJob(context.Background(), job.ID)
		require.NoError(t, err)

		job, err = db.GetJob(context.Background())
		require.NoError(t, err)
		require.Nil(t, job)
	})
}
