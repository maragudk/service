package s3_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/maragudk/service/s3test"
)

func TestBlobStore(t *testing.T) {
	s3test.SkipIfShort(t)

	t.Run("puts, gets, and deletes a blob", func(t *testing.T) {
		objectStore := s3test.CreateObjectStore(t)

		err := objectStore.Put(context.Background(), s3test.DefaultBucket, "test", "text/plain",
			strings.NewReader("hello"))
		require.NoError(t, err)

		body, err := objectStore.Get(context.Background(), s3test.DefaultBucket, "test")
		require.NoError(t, err)
		bodyBytes, err := io.ReadAll(body)
		require.NoError(t, err)
		require.Equal(t, "hello", string(bodyBytes))

		err = objectStore.Delete(context.Background(), s3test.DefaultBucket, "test")
		require.NoError(t, err)

		body, err = objectStore.Get(context.Background(), s3test.DefaultBucket, "test")
		require.NoError(t, err)
		require.Nil(t, body)
	})
}
