package billing

import (
	"log/slog"

	"github.com/basilex/promenade/pkg/module"
)

func init() {
	// Auto-register billing module when package is imported
	mod := New()
	if err := module.DefaultRegistry.Register(mod); err != nil {
		slog.Error("Failed to register billing module", "error", err)
	}
}
