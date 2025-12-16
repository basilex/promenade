package uuidv7

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNew(t *testing.T) {
	u := New()

	// Check it's a valid UUID
	if u == uuid.Nil {
		t.Fatal("generated UUID is nil")
	}

	// Check version is 7
	if !IsV7(u) {
		t.Fatalf("UUID version is not 7: got %d", u[6]>>4)
	}

	// Check variant is RFC 4122
	variant := (u[8] >> 6) & 0x3
	if variant != 0x2 { // 10 in binary
		t.Fatalf("UUID variant is incorrect: got %b", variant)
	}
}

func TestNewWithTime(t *testing.T) {
	now := time.Now()
	u := NewWithTime(now)

	// Extract timestamp and compare (allow 1ms tolerance)
	extractedTime := ExtractTime(u)
	diff := extractedTime.Sub(now).Abs()

	if diff > time.Millisecond {
		t.Fatalf("timestamp mismatch: expected %v, got %v (diff: %v)",
			now, extractedTime, diff)
	}
}

func TestExtractTime(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 12, 30, 45, 123000000, time.UTC)
	u := NewWithTime(testTime)

	extracted := ExtractTime(u)

	// Check millisecond precision
	if extracted.UnixMilli() != testTime.UnixMilli() {
		t.Fatalf("extracted time mismatch: expected %d ms, got %d ms",
			testTime.UnixMilli(), extracted.UnixMilli())
	}
}

func TestExtractTimeInvalidVersion(t *testing.T) {
	// Create a UUID v4 (random)
	v4 := uuid.New()

	extracted := ExtractTime(v4)

	// Should return zero time for non-v7 UUIDs
	if !extracted.IsZero() {
		t.Fatal("ExtractTime should return zero time for non-v7 UUID")
	}
}

func TestIsV7(t *testing.T) {
	v7 := New()
	v4 := uuid.New()

	if !IsV7(v7) {
		t.Fatal("IsV7 returned false for v7 UUID")
	}

	if IsV7(v4) {
		t.Fatal("IsV7 returned true for v4 UUID")
	}
}

func TestTimeOrdering(t *testing.T) {
	// Generate multiple UUIDs with increasing timestamps
	times := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC),
		time.Date(2024, 12, 31, 23, 59, 59, 999000000, time.UTC),
	}

	var uuids []uuid.UUID
	for _, t := range times {
		uuids = append(uuids, NewWithTime(t))
	}

	// Verify lexicographic ordering matches time ordering
	for i := 1; i < len(uuids); i++ {
		prev := uuids[i-1].String()
		curr := uuids[i].String()

		if prev >= curr {
			t.Fatalf("UUID ordering broken: %s should be less than %s",
				prev, curr)
		}
	}
}

func TestUniqueness(t *testing.T) {
	const count = 10000
	seen := make(map[uuid.UUID]bool, count)

	for i := 0; i < count; i++ {
		u := New()
		if seen[u] {
			t.Fatalf("duplicate UUID generated: %s", u)
		}
		seen[u] = true
	}
}

// Benchmark UUID v7 generation
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New()
	}
}

// Benchmark UUID v4 generation for comparison
func BenchmarkNewV4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = uuid.New()
	}
}

// Benchmark time extraction
func BenchmarkExtractTime(b *testing.B) {
	u := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ExtractTime(u)
	}
}
