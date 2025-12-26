package workflows

import (
	"log/slog"

	"github.com/basilex/promenade/pkg/module"
)

func init() {
	if err := module.DefaultRegistry.Register(NewModule()); err != nil {
		slog.Error("Failed to register workflows module", "error", err)
	}
}
