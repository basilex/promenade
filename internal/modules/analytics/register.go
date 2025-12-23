package analytics

import (
	"log/slog"

	"github.com/basilex/promenade/pkg/module"
)

func init() {
	mod := New()
	if err := module.DefaultRegistry.Register(mod); err != nil {
		slog.Error("Failed to register analytics module", "error", err)
	}
}
