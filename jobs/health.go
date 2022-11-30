package jobs

import (
	"context"

	"github.com/maragudk/service/model"
)

func Health(r registry) {
	r.Register("health", func(ctx context.Context, m model.Map) error {
		return nil
	})
}
