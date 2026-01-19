package scripting

import (
	"context"
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// Engine is the main LUA scripting engine
type Engine struct {
	config  Config
	stdlib  *StandardLibrary
	sandbox *Sandbox
}

// Config holds engine configuration
type Config struct {
	MemoryLimit   int64         // Memory limit in bytes
	Timeout       time.Duration // CPU timeout
	MaxGoroutines int           // Max goroutines allowed
	AllowFileIO   bool          // Enable filesystem access
	AllowNetwork  bool          // Enable network access
	Debug         bool          // Enable debug mode
}

// DefaultConfig returns default engine configuration
func DefaultConfig() Config {
	return Config{
		MemoryLimit:   50 * 1024 * 1024, // 50MB
		Timeout:       5 * time.Second,
		MaxGoroutines: 0,
		AllowFileIO:   false,
		AllowNetwork:  false,
		Debug:         false,
	}
}

// NewEngine creates a new LUA engine with configuration and standard library
func NewEngine(config Config, stdlib *StandardLibrary) *Engine {
	return &Engine{
		config:  config,
		stdlib:  stdlib,
		sandbox: NewSandbox(config),
	}
}

// Execute executes a LUA script with optional parameters
func (e *Engine) Execute(ctx context.Context, script string, params map[string]interface{}) (interface{}, error) {
	// Execute with timeout
	resultCh := make(chan interface{}, 1)
	errorCh := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorCh <- fmt.Errorf("script panic: %v", r)
			}
		}()

		L := lua.NewState()
		defer L.Close()

		// Apply security sandbox
		e.sandbox.Apply(L)

		// Register standard library
		e.stdlib.Register(L)

		// Set parameters
		for k, v := range params {
			L.SetGlobal(k, e.convertGoToLua(L, v))
		}

		// Execute script
		if err := L.DoString(script); err != nil {
			errorCh <- fmt.Errorf("script execution failed: %w", err)
			return
		}

		// Get return value from stack
		if L.GetTop() > 0 {
			result := e.convertLuaValue(L.Get(-1))
			resultCh <- result
		} else {
			resultCh <- nil
		}
	}()

	select {
	case result := <-resultCh:
		return result, nil
	case err := <-errorCh:
		return nil, err
	case <-time.After(e.config.Timeout):
		return nil, fmt.Errorf("script execution timeout after %v", e.config.Timeout)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// ExecuteFunction executes a specific LUA function with arguments
func (e *Engine) ExecuteFunction(ctx context.Context, script string, functionName string, args ...interface{}) (interface{}, error) {
	// Execute with timeout
	resultCh := make(chan interface{}, 1)
	errorCh := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errorCh <- fmt.Errorf("function panic: %v", r)
			}
		}()

		L := lua.NewState()
		defer L.Close()

		// Apply security sandbox
		e.sandbox.Apply(L)

		// Register standard library
		e.stdlib.Register(L)

		// Compile script
		if err := L.DoString(script); err != nil {
			errorCh <- fmt.Errorf("script compilation failed: %w", err)
			return
		}

		// Get function
		fn := L.GetGlobal(functionName)
		if fn.Type() != lua.LTFunction {
			errorCh <- fmt.Errorf("function %s not found", functionName)
			return
		}

		// Push function and arguments
		L.Push(fn)
		for _, arg := range args {
			L.Push(e.convertGoToLua(L, arg))
		}

		// Call function
		if err := L.PCall(len(args), 1, nil); err != nil {
			errorCh <- fmt.Errorf("function execution failed: %w", err)
			return
		}

		// Get return value
		if L.GetTop() > 0 {
			result := e.convertLuaValue(L.Get(-1))
			resultCh <- result
		} else {
			resultCh <- nil
		}
	}()

	select {
	case result := <-resultCh:
		return result, nil
	case err := <-errorCh:
		return nil, err
	case <-time.After(e.config.Timeout):
		return nil, fmt.Errorf("function execution timeout after %v", e.config.Timeout)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Validate checks if a script has valid syntax
func (e *Engine) Validate(script string) error {
	L := lua.NewState()
	defer L.Close()

	_, err := L.LoadString(script)
	return err
}

// convertLuaValue converts LUA value to Go interface{}
func (e *Engine) convertLuaValue(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		result := make(map[string]interface{})
		v.ForEach(func(key lua.LValue, value lua.LValue) {
			if keyStr, ok := key.(lua.LString); ok {
				result[string(keyStr)] = e.convertLuaValue(value)
			}
		})
		return result
	default:
		return lv.String()
	}
}

// convertGoToLua converts Go interface{} to LUA value
func (e *Engine) convertGoToLua(L *lua.LState, val interface{}) lua.LValue {
	if val == nil {
		return lua.LNil
	}

	switch v := val.(type) {
	case bool:
		return lua.LBool(v)
	case int:
		return lua.LNumber(v)
	case int64:
		return lua.LNumber(v)
	case float64:
		return lua.LNumber(v)
	case string:
		return lua.LString(v)
	case map[string]interface{}:
		table := L.NewTable()
		for k, v := range v {
			table.RawSetString(k, e.convertGoToLua(L, v))
		}
		return table
	case []interface{}:
		table := L.NewTable()
		for i, v := range v {
			table.RawSetInt(i+1, e.convertGoToLua(L, v)) // LUA arrays are 1-indexed
		}
		return table
	default:
		// For unsupported types, convert to string
		return lua.LString(fmt.Sprintf("%v", v))
	}
}
