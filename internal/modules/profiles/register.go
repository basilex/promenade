package profiles

import (
	"log/slog"

	"github.com/basilex/promenade/pkg/module"
)

func init() {
	// Auto-register profiles module when package is imported
	mod := New()
	if err := module.DefaultRegistry.Register(mod); err != nil {
		slog.Error("Failed to register profiles module", "error", err)
	}
}
