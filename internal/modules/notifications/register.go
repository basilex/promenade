package notifications

import (
	"log/slog"

	"github.com/basilex/promenade/pkg/module"
)

func init() {
	if err := module.DefaultRegistry.Register(New()); err != nil {
		slog.Error("Failed to register notifications module", "error", err)
	}
}
