# Session 10: User Aggregate - Entity Tests Refactoring

**Date**: January 11, 2026  
**Context**: Sales Domain (5/10 contexts)  
**Pattern**: Test-Only Refactoring (entity.go already clean)  
**Duration**: ~30 minutes (10 implementation + 20 debugging)

---

## Summary

User context represents the first **test-only refactoring** in Phase 2 - where entity.go is already clean (0 fmt.Errorf), requiring only test assertion updates.

**Key Achievements**:
- **Test-only changes**: entity.go already production-ready
- **9 string assertions** → errors.Is() pattern
- **2 bugs discovered and fixed**: Missing import, wrong error constant
- **Cross-layer error pattern discovered**: valueobject uses fmt.Errorf, domain wraps with constants
- **All tests passing**: 100% success rate (entity + usecase tests)

---

## Metrics

| Metric | Value |
|--------|-------|
| **fmt.Errorf Eliminated** | 0 (entity.go already clean) |
| **New Error Constants** | 0 (28 existing sufficient) |
| **Test Assertions Changed** | 9 (string → errors.Is) |
| **Bugs Fixed** | 2 (import + error constant) |
| **Total Tests** | All passing ✅ |
| **Pass Rate** | 100% |
| **Duration** | ~30 minutes |

---

## Implementation Details

### Phase 1: Initial Discovery (Operations 120-122)

**Operation 120: Read errors.go** ✅
```bash
# Discovered 28 existing domain errors
Lines: 1-180

Error Categories:
- Authentication: 7 errors (ErrEmailRequired, ErrPasswordRequired, etc.)
- Password Validation: 6 errors (ErrPasswordTooShort, ErrPasswordRequiresDigit, etc.)
- User Validation: 4 errors (ErrEmailRequired, ErrPasswordHashRequired, etc.)
- Account Management: 11 errors (ErrUserNotFound, ErrEmailAlreadyExists, etc.)

Key Finding: Comprehensive error coverage, no additions needed
```

**Operation 121: Scan entity.go** ✅
```bash
# Search for fmt.Errorf in entity.go
Result: 0 matches

Analysis: entity.go already uses domain error constants exclusively
Pattern: Quality compounds - Phase 1 work already complete
Conclusion: Test-only refactoring (first occurrence in Phase 2)
```

**Operation 122: Scan entity_test.go** ✅
```bash
# Search for string assertions
Result: 9 matches found

Locations:
- Line 36:  assert.Contains(t, err.Error(), "invalid email")
- Line 42:  assert.Contains(t, err.Error(), "password must be at least")
- Line 48:  assert.Contains(t, err.Error(), "password must contain at least one digit")
- Line 54:  assert.Contains(t, err.Error(), "password must contain at least one letter")
- Line 68:  assert.Contains(t, err.Error(), "password too long")
- Line 106: assert.Contains(t, err.Error(), "password must be at least")
- Line 300: assert.Contains(t, err.Error(), "email is required")
- Line 313: assert.Contains(t, err.Error(), "password hash is required")
- Line 326: assert.Contains(t, err.Error(), "invalid user status")

Analysis: All assertions map cleanly to existing domain errors
```

---

### Phase 2: Implementation (Operations 123-126)

**Operation 123-125: Context Reading** ✅
```bash
# Read all 9 assertion contexts for accurate replacement

Key Findings:
- All assertions in straightforward test cases
- Clear 1:1 mapping to domain errors
- No complex multi-error scenarios
- Standard testify/assert pattern throughout
```

**Operation 126: Multi-Replace Execution** ✅
```bash
# Execute all 9 replacements via multi_replace_string_in_file

Changes Applied:
1. Line 36:  ErrInvalidEmailFormat (later fixed to correct error)
2. Line 42:  ErrPasswordTooShort
3. Line 48:  ErrPasswordRequiresDigit
4. Line 54:  ErrPasswordRequiresLetter
5. Line 68:  ErrPasswordTooLong
6. Line 106: ErrPasswordTooShort
7. Line 300: ErrEmailRequired
8. Line 313: ErrPasswordHashRequired
9. Line 326: ErrInvalidUserStatus

Result: All 9 replacements successful
Status: Implementation phase complete
```

---

### Phase 3: Bug Discovery & Fixes (Operations 127-138)

#### Bug 1: Missing errors Import (Operations 127-134)

**Operations 127-131: Test Execution Attempts** ❌
```bash
# Multiple attempts to run tests failed due to terminal issues
Attempts:
- go test -v .
- go test -v -count=1 .
- go test from project root
- go test -c (compilation check)
- go test with piping

Issue: Terminal output problems prevented seeing errors
Lesson: Plain go test unreliable in this setup
```

**Operation 132: make test-unit Discovery** ✅ CRITICAL
```bash
Command: make test-unit 2>&1 | grep -A 20 "identity/user"

Output:
# github.com/basilex/promenade/internal/contexts/identity/user
internal/contexts/identity/user/entity_test.go:36:18: undefined: errors
internal/contexts/identity/user/entity_test.go:36:45: undefined: valueobject.ErrInvalidEmail
internal/contexts/identity/user/entity_test.go:42:18: undefined: errors
internal/contexts/identity/user/entity_test.go:48:18: undefined: errors
internal/contexts/identity/user/entity_test.go:54:18: undefined: errors
internal/contexts/identity/user/entity_test.go:68:18: undefined: errors
internal/contexts/identity/user/entity_test.go:106:18: undefined: errors
internal/contexts/identity/user/entity_test.go:300:18: undefined: errors
internal/contexts/identity/user/entity_test.go:313:18: undefined: errors
internal/contexts/identity/user/entity_test.go:326:18: undefined: errors
FAIL    github.com/basilex/promenade/internal/contexts/identity/user [build failed]

Bugs Found:
1. Missing "errors" import (9 instances)
2. Non-existent valueobject.ErrInvalidEmail (1 instance)

Key Insight: make test-unit more reliable than plain go test
```

**Operations 133-134: Fix Missing Import** ✅
```go
// Operation 133: Read current imports
// Found: "errors" package missing

// Operation 134: Add errors import
Before:
package user

import (
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

After:
package user

import (
	"errors"  // ← ADDED for errors.Is() support
	"testing"
	"time"

	"github.com/basilex/promenade/pkg/uuidv7"
	"github.com/basilex/promenade/pkg/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

Result: ✅ Fixed 9 "undefined: errors" compilation errors
```

#### Bug 2: Wrong Error Constant (Operations 135-138)

**Operation 135: Test After Import Fix** ⚠️
```bash
Command: go test ./internal/contexts/identity/user -v

Output:
# github.com/basilex/promenade/internal/contexts/identity/user
internal/contexts/identity/user/entity_test.go:37:45: undefined: valueobject.ErrInvalidEmail
FAIL    github.com/basilex/promenade/internal/contexts/identity/user [build failed]

Analysis: Import fix successful, but second bug revealed
```

**Operation 136: Search valueobject for Error** ✅
```bash
Command: grep_search for "^var Err.*Email.*=" in pkg/valueobject/email.go

Result: No matches found

Analysis: valueobject package has NO error constants defined
Question: What error pattern does valueobject use?
```

**Operation 137: Read valueobject/email.go** ✅ CRITICAL DISCOVERY
```go
// Discovered: valueobject uses fmt.Errorf (not error constants)

func NewEmail(address string) (Email, error) {
	address = strings.TrimSpace(strings.ToLower(address))

	if address == "" {
		return Email{}, fmt.Errorf("email address is required")
	}

	if len(address) > 254 {
		return Email{}, fmt.Errorf("email address too long (max 254 characters)")
	}

	if !emailRegex.MatchString(address) {
		return Email{}, fmt.Errorf("invalid email format: %s", address)
	}

	return Email{value: address}, nil
}

Key Findings:
- valueobject package uses fmt.Errorf (infrastructure layer)
- No error constants defined in valueobject
- Cannot reference valueobject.ErrInvalidEmail (doesn't exist)
```

**Operation 137b: Check entity.go Error Handling** ✅ PATTERN FOUND
```go
// How does entity.go handle valueobject errors?

func NewUser(email, password string) (*User, error) {
	// Validate email
	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, errors.Join(ErrInvalidEmailFormat, err)
	}
	// ...
}

PATTERN DISCOVERED:
- entity.go calls valueobject.NewEmail()
- If error occurs, wraps with errors.Join(ErrInvalidEmailFormat, err)
- ErrInvalidEmailFormat is the domain error (user package)
- Domain layer provides error abstraction over infrastructure errors

SOLUTION:
- Test should check: ErrInvalidEmailFormat (domain error)
- Not: valueobject.ErrInvalidEmail (doesn't exist)
- Reason: Domain layer owns error contracts
```

**Operation 138: Fix Error Constant** ✅
```go
// Replace non-existent error with correct domain error

Before:
t.Run("empty email", func(t *testing.T) {
	_, err := NewUser("", "Password123")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, valueobject.ErrInvalidEmail))  // ❌ WRONG
})

After:
t.Run("empty email", func(t *testing.T) {
	_, err := NewUser("", "Password123")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidEmailFormat))  // ✅ CORRECT
})

Result: ✅ Fixed final compilation error
Pattern: Domain errors are the public contract
```

---

## Complete Replacements

### All 9 Test Assertions (Final State)

**Replacement 1: Invalid Email (Line 36)** ✅ FIXED
```go
// Before
assert.Contains(t, err.Error(), "invalid email")

// After (initial - BUG)
assert.True(t, errors.Is(err, valueobject.ErrInvalidEmail))

// After (final - FIXED)
assert.True(t, errors.Is(err, ErrInvalidEmailFormat))
```

**Replacement 2: Password Too Short (Line 42)** ✅
```go
// Before
assert.Contains(t, err.Error(), "password must be at least")

// After
assert.True(t, errors.Is(err, ErrPasswordTooShort))
```

**Replacement 3: Password Requires Digit (Line 48)** ✅
```go
// Before
assert.Contains(t, err.Error(), "password must contain at least one digit")

// After
assert.True(t, errors.Is(err, ErrPasswordRequiresDigit))
```

**Replacement 4: Password Requires Letter (Line 54)** ✅
```go
// Before
assert.Contains(t, err.Error(), "password must contain at least one letter")

// After
assert.True(t, errors.Is(err, ErrPasswordRequiresLetter))
```

**Replacement 5: Password Too Long (Line 68)** ✅
```go
// Before
assert.Contains(t, err.Error(), "password too long")

// After
assert.True(t, errors.Is(err, ErrPasswordTooLong))
```

**Replacement 6: Password Too Short (Line 106)** ✅
```go
// Before
assert.Contains(t, err.Error(), "password must be at least")

// After
assert.True(t, errors.Is(err, ErrPasswordTooShort))
```

**Replacement 7: Email Required (Line 300)** ✅
```go
// Before
assert.Contains(t, err.Error(), "email is required")

// After
assert.True(t, errors.Is(err, ErrEmailRequired))
```

**Replacement 8: Password Hash Required (Line 313)** ✅
```go
// Before
assert.Contains(t, err.Error(), "password hash is required")

// After
assert.True(t, errors.Is(err, ErrPasswordHashRequired))
```

**Replacement 9: Invalid User Status (Line 326)** ✅
```go
// Before
assert.Contains(t, err.Error(), "invalid user status")

// After
assert.True(t, errors.Is(err, ErrInvalidUserStatus))
```

---

## Cross-Layer Error Handling Pattern

### Architecture Discovery

**Infrastructure Layer** (valueobject):
```go
// Uses fmt.Errorf (simple error strings)
func NewEmail(address string) (Email, error) {
	if address == "" {
		return Email{}, fmt.Errorf("email address is required")
	}
	if !emailRegex.MatchString(address) {
		return Email{}, fmt.Errorf("invalid email format: %s", address)
	}
	return Email{value: address}, nil
}

Characteristics:
- No error constants defined
- Uses fmt.Errorf directly
- Returns basic errors
- Implementation details exposed
```

**Domain Layer** (user entity):
```go
// Wraps infrastructure errors with domain constants
func NewUser(email, password string) (*User, error) {
	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, errors.Join(ErrInvalidEmailFormat, err)
	}
	// ...
}

Characteristics:
- Defines domain error constants
- Wraps infrastructure errors
- Provides error abstraction
- Public error contract
```

**Test Layer**:
```go
// Checks domain errors (public contracts)
t.Run("empty email", func(t *testing.T) {
	_, err := NewUser("", "Password123")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidEmailFormat))
})

Characteristics:
- Uses errors.Is() to check error chain
- Checks domain errors (not infrastructure)
- Ignores valueobject implementation details
- Tests public contracts
```

### Error Wrapping Flow

```
User Input (empty email)
    ↓
entity.NewUser("", "Password123")
    ↓
valueobject.NewEmail("")
    ↓
fmt.Errorf("email address is required")  ← Infrastructure error
    ↓
errors.Join(ErrInvalidEmailFormat, err)  ← Domain wrapping
    ↓
return nil, wrappedError
    ↓
Test: errors.Is(err, ErrInvalidEmailFormat) → true  ← Domain check
```

### Key Architectural Insights

1. **Layered Error Handling**:
   - Infrastructure: fmt.Errorf (simple strings)
   - Domain: Error constants + wrapping
   - Tests: Domain error checking

2. **Error Abstraction**:
   - Domain layer hides infrastructure details
   - Provides stable error contracts
   - Consumers check domain errors only

3. **errors.Join Pattern**:
   - Preserves error chain
   - errors.Is() checks full chain
   - Works with wrapped errors

4. **Why This Matters**:
   - valueobject can change error messages
   - Domain contract remains stable
   - Tests don't break on infrastructure changes
   - Clear separation of concerns

---

## Key Lessons Learned

### 1. Test-Only Refactoring Pattern
**Discovery**: entity.go already clean (0 fmt.Errorf)  
**Implication**: Quality compounds - Phase 1 work pays off  
**Pattern**: First occurrence in Phase 2  
**Benefit**: Faster refactoring (only test changes)

### 2. Cross-Layer Error Handling
**Discovery**: valueobject uses fmt.Errorf, domain wraps with constants  
**Pattern**: Infrastructure → Domain → Tests  
**Lesson**: Cannot assume error constants exist in other packages  
**Solution**: Domain layer provides error abstraction

### 3. Import Management
**Bug**: Forgot "errors" import when using errors.Is()  
**Impact**: 9 compilation errors  
**Lesson**: Always verify imports when adding new dependencies  
**Prevention**: Add import verification to workflow

### 4. Error Constant Validation
**Bug**: Referenced non-existent valueobject.ErrInvalidEmail  
**Impact**: 1 compilation error  
**Lesson**: Verify error constants exist before using  
**Prevention**: Check cross-package error references

### 5. make test-unit Reliability
**Problem**: Plain go test failed silently (terminal issues)  
**Solution**: Use make test-unit for reliable error detection  
**Benefit**: Reveals compilation errors consistently  
**Adoption**: Use make test-unit as standard practice

### 6. Error Wrapping with errors.Join
**Pattern**: errors.Join(DomainError, infrastructureErr)  
**Benefit**: Preserves error chain, stable contracts  
**Usage**: errors.Is() checks full chain  
**Architecture**: Domain abstracts infrastructure

---

## Bug Prevention Checklist

**For Future Contexts**:
- [ ] Verify all error constants exist before using
- [ ] Add "errors" import when using errors.Is()
- [ ] Check for cross-package error references
- [ ] Use make test-unit instead of plain go test
- [ ] Verify compilation before running tests
- [ ] Document cross-layer error patterns
- [ ] Test error.Is() with wrapped errors

---

## Test Results

**Final Test Execution** ✅
```bash
Command: go test ./internal/contexts/identity/user -v

Results:
=== RUN   TestNewUser
=== RUN   TestNewUser/valid_user
=== RUN   TestNewUser/empty_email
=== RUN   TestNewUser/invalid_password
=== RUN   TestNewUser/password_without_digit
=== RUN   TestNewUser/password_without_letter
=== RUN   TestNewUser/password_too_long
--- PASS: TestNewUser (0.07s)

[... all other tests passing ...]

PASS
ok      github.com/basilex/promenade/internal/contexts/identity/user    2.616s

All entity and usecase tests: ✅ PASS
All 9 errors.Is() assertions: ✅ Working correctly
ErrInvalidEmailFormat: ✅ Properly checking email validation
All password validations: ✅ Working correctly
No regressions: ✅ Confirmed
```

---

## Status

**User Context**: ✅ 100% COMPLETE

**Files Modified**: 1
- `internal/contexts/identity/user/entity_test.go` (9 assertions + 1 import + 1 error fix)

**Files Unchanged**: 2
- `internal/contexts/identity/user/errors.go` (28 errors sufficient)
- `internal/contexts/identity/user/entity.go` (0 fmt.Errorf - already clean)

**Bugs Fixed**: 2
- Missing "errors" import
- Wrong error constant (valueobject.ErrInvalidEmail → ErrInvalidEmailFormat)

**Tests Status**: 100% passing  
**Duration**: ~30 minutes (10 implementation + 20 debugging)  
**Next Context**: Deal (6/10)

---

## Architectural Impact

**Pattern Established**: Test-Only Refactoring
- Demonstrates value of Phase 1 investment
- Faster refactoring when entity.go already clean
- Focus only on test modernization

**Cross-Layer Error Handling Documented**:
- Infrastructure (valueobject): fmt.Errorf
- Domain (entity): Error constants + wrapping
- Tests: Domain error checking

**Error Abstraction Benefits**:
- Stable public contracts
- Infrastructure can change without breaking tests
- Clear separation of concerns
- Type-safe error handling

**Lessons Applied to Future Contexts**:
- Import verification required
- Error constant existence checks
- make test-unit as standard
- Cross-package error validation

---

**Completed**: January 11, 2026  
**Next Session**: Deal context (6/10) - Sales domain continuation
