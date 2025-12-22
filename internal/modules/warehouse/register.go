package warehouse

import (
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/module"
)

func init() {
	// Auto-register warehouse module when package is imported
	// Note: This is a commercial module and requires license verification
	mod := New()
	if err := module.DefaultRegistry.Register(mod); err != nil {
		logger.Fatal("Failed to register warehouse module", "error", err)
	}
}
