package jobs

import (
	"context"
	"log"

	"github.com/maragudk/service/model"
)

func health(r registry, log *log.Logger) {
	r.Register("health", func(ctx context.Context, m model.Map) error {
		return nil
	})
}
