package scripting

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewEngine(t *testing.T) {
	config := DefaultConfig()
	engine := NewEngine(config)
	
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.stdlib)
	assert.NotNil(t, engine.sandbox)
}

func TestEngine_Execute_SimpleScript(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	ctx := context.Background()
	
	script := `return 2 + 2`
	
	result, err := engine.Execute(ctx, script, nil)
	
	assert.NoError(t, err)
	assert.Equal(t, float64(4), result)
}

func TestEngine_Execute_WithParameters(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	ctx := context.Background()
	
	script := `return x + y`
	params := map[string]interface{}{
		"x": 10,
		"y": 20,
	}
	
	result, err := engine.Execute(ctx, script, params)
	
	assert.NoError(t, err)
	assert.Equal(t, float64(30), result)
}

func TestEngine_Execute_StringResult(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	ctx := context.Background()
	
	script := `return "Hello, World!"`
	
	result, err := engine.Execute(ctx, script, nil)
	
	assert.NoError(t, err)
	assert.Equal(t, "Hello, World!", result)
}

func TestEngine_Execute_BooleanResult(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	ctx := context.Background()
	
	script := `return true`
	
	result, err := engine.Execute(ctx, script, nil)
	
	assert.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestEngine_Execute_TableResult(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	ctx := context.Background()
	
	script := `return {name = "John", age = 30}`
	
	result, err := engine.Execute(ctx, script, nil)
	
	assert.NoError(t, err)
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "John", resultMap["name"])
	assert.Equal(t, float64(30), resultMap["age"])
}

func TestEngine_Execute_SyntaxError(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	ctx := context.Background()
	
	script := `return 2 +`  // Syntax error
	
	_, err := engine.Execute(ctx, script, nil)
	
	assert.Error(t, err)
}

func TestEngine_Execute_Timeout(t *testing.T) {
	config := DefaultConfig()
	config.Timeout = 100 * time.Millisecond
	engine := NewEngine(config)
	ctx := context.Background()
	
	script := `while true do end`  // Infinite loop
	
	_, err := engine.Execute(ctx, script, nil)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestEngine_ExecuteFunction_Success(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	ctx := context.Background()
	
	script := `
		function greet(name)
			return "Hello, " .. name .. "!"
		end
	`
	
	result, err := engine.ExecuteFunction(ctx, script, "greet", "Alice")
	
	assert.NoError(t, err)
	assert.Equal(t, "Hello, Alice!", result)
}

func TestEngine_ExecuteFunction_NotFound(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	ctx := context.Background()
	
	script := `function foo() end`
	
	_, err := engine.ExecuteFunction(ctx, script, "bar")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEngine_Validate_Success(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	
	script := `return 2 + 2`
	
	err := engine.Validate(script)
	
	assert.NoError(t, err)
}

func TestEngine_Validate_SyntaxError(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	
	script := `return 2 +`
	
	err := engine.Validate(script)
	
	assert.Error(t, err)
}

func TestEngine_StandardLibrary(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	ctx := context.Background()
	
	script := `return Customer.GetTier("customer-123")`
	
	result, err := engine.Execute(ctx, script, nil)
	
	assert.NoError(t, err)
	assert.Equal(t, "free", result)
}

func TestEngine_Sandbox_NoDangerousFunctions(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	ctx := context.Background()
	
	tests := []struct {
		name     string
		script   string
	}{
		{"dofile", `return type(dofile)`},
		{"loadfile", `return type(loadfile)`},
		{"require", `return type(require)`},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Execute(ctx, tt.script, nil)
			assert.NoError(t, err)
			assert.Equal(t, "nil", result)
		})
	}
}

func TestEngine_Sandbox_NoFileIO(t *testing.T) {
	engine := NewEngine(DefaultConfig())
	ctx := context.Background()
	
	script := `return type(io)`
	
	result, err := engine.Execute(ctx, script, nil)
	
	assert.NoError(t, err)
	assert.Equal(t, "nil", result)
}

// Benchmarks
func BenchmarkEngine_Execute_Simple(b *testing.B) {
	engine := NewEngine(DefaultConfig())
	ctx := context.Background()
	script := `return 2 + 2`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Execute(ctx, script, nil)
	}
}

func BenchmarkEngine_Execute_WithParams(b *testing.B) {
	engine := NewEngine(DefaultConfig())
	ctx := context.Background()
	script := `return x + y`
	params := map[string]interface{}{"x": 10, "y": 20}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Execute(ctx, script, params)
	}
}

func BenchmarkEngine_ExecuteFunction(b *testing.B) {
	engine := NewEngine(DefaultConfig())
	ctx := context.Background()
	script := `function add(a, b) return a + b end`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.ExecuteFunction(ctx, script, "add", 10, 20)
	}
}
