package scripting

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// Sandbox provides security restrictions for LUA VM
type Sandbox struct {
	config Config
}

// NewSandbox creates a new sandbox with configuration
func NewSandbox(config Config) *Sandbox {
	return &Sandbox{config: config}
}

// Apply applies all security restrictions to LUA state
func (s *Sandbox) Apply(L *lua.LState) {
	s.restrictDangerousFunctions(L)

	if !s.config.AllowFileIO {
		s.restrictFileIO(L)
	}

	if !s.config.AllowNetwork {
		s.restrictNetwork(L)
	}
}

// restrictDangerousFunctions removes dangerous LUA functions
func (s *Sandbox) restrictDangerousFunctions(L *lua.LState) {
	// Remove dynamic code loading functions
	L.SetGlobal("dofile", lua.LNil)
	L.SetGlobal("loadfile", lua.LNil)
	L.SetGlobal("load", lua.LNil)
	L.SetGlobal("loadstring", lua.LNil)

	// Remove module system
	L.SetGlobal("require", lua.LNil)
	L.SetGlobal("module", lua.LNil)

	// Remove environment manipulation
	L.SetGlobal("setfenv", lua.LNil)
	L.SetGlobal("getfenv", lua.LNil)
}

// restrictFileIO removes filesystem access
func (s *Sandbox) restrictFileIO(L *lua.LState) {
	// Remove io library completely
	L.SetGlobal("io", lua.LNil)

	// Remove dangerous os functions
	os := L.GetGlobal("os")
	if osTable, ok := os.(*lua.LTable); ok {
		osTable.RawSetString("execute", lua.LNil)
		osTable.RawSetString("exit", lua.LNil)
		osTable.RawSetString("remove", lua.LNil)
		osTable.RawSetString("rename", lua.LNil)
		osTable.RawSetString("tmpname", lua.LNil)
		osTable.RawSetString("getenv", lua.LNil)
		osTable.RawSetString("setlocale", lua.LNil)
	}
}

// restrictNetwork removes network access
func (s *Sandbox) restrictNetwork(L *lua.LState) {
	// Remove socket library (if loaded)
	L.SetGlobal("socket", lua.LNil)

	// Remove http libraries
	L.SetGlobal("http", lua.LNil)
	L.SetGlobal("https", lua.LNil)
	L.SetGlobal("ftp", lua.LNil)
}

// SetMemoryLimit sets memory limit for LUA VM
func (s *Sandbox) SetMemoryLimit(L *lua.LState) error {
	// TODO: Implement memory limiting
	// Options:
	// 1. Track allocations via custom allocator
	// 2. Use OS-level memory limits (cgroups on Linux)
	// 3. Periodic memory checks with runtime.ReadMemStats()
	return fmt.Errorf("memory limiting not yet implemented")
}
