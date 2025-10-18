package retry

import (
	"bytes"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
)

// TestDatabasePolicy tests the database-specific retry policy
func TestDatabasePolicy(t *testing.T) {
	t.Parallel()
	policy := DatabasePolicy()

	if policy.MaxRetries != 5 {
		t.Errorf("DatabasePolicy: expected MaxRetries=5, got %d", policy.MaxRetries)
	}

	if policy.InitialDelay != 50*time.Millisecond {
		t.Errorf("DatabasePolicy: expected InitialDelay=50ms, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 2*time.Second {
		t.Errorf("DatabasePolicy: expected MaxDelay=2s, got %v", policy.MaxDelay)
	}

	if policy.Multiplier != 2.0 {
		t.Errorf("DatabasePolicy: expected Multiplier=2.0, got %f", policy.Multiplier)
	}

	if !policy.Jitter {
		t.Error("DatabasePolicy: expected Jitter=true")
	}
}

// TestHTTPPolicy tests the HTTP-specific retry policy
func TestHTTPPolicy(t *testing.T) {
	t.Parallel()
	policy := HTTPPolicy()

	if policy.MaxRetries != 3 {
		t.Errorf("HTTPPolicy: expected MaxRetries=3, got %d", policy.MaxRetries)
	}

	if policy.InitialDelay != 100*time.Millisecond {
		t.Errorf("HTTPPolicy: expected InitialDelay=100ms, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 5*time.Second {
		t.Errorf("HTTPPolicy: expected MaxDelay=5s, got %v", policy.MaxDelay)
	}

	if policy.Multiplier != 2.0 {
		t.Errorf("HTTPPolicy: expected Multiplier=2.0, got %f", policy.Multiplier)
	}

	if !policy.Jitter {
		t.Error("HTTPPolicy: expected Jitter=true")
	}
}

// TestFileIOPolicy tests the file I/O specific retry policy
func TestFileIOPolicy(t *testing.T) {
	t.Parallel()
	policy := FileIOPolicy()

	if policy.MaxRetries != 2 {
		t.Errorf("FileIOPolicy: expected MaxRetries=2, got %d", policy.MaxRetries)
	}

	if policy.InitialDelay != 10*time.Millisecond {
		t.Errorf("FileIOPolicy: expected InitialDelay=10ms, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 500*time.Millisecond {
		t.Errorf("FileIOPolicy: expected MaxDelay=500ms, got %v", policy.MaxDelay)
	}

	if policy.Multiplier != 2.0 {
		t.Errorf("FileIOPolicy: expected Multiplier=2.0, got %f", policy.Multiplier)
	}

	if !policy.Jitter {
		t.Error("FileIOPolicy: expected Jitter=true")
	}
}

// TestPolicyBuilder_Default tests the default policy builder
func TestPolicyBuilder_Default(t *testing.T) {
	t.Parallel()
	builder := NewPolicyBuilder()
	policy := builder.Build()

	// Should have default values
	if policy.MaxRetries != 3 {
		t.Errorf("Default PolicyBuilder: expected MaxRetries=3, got %d", policy.MaxRetries)
	}

	if policy.InitialDelay != 100*time.Millisecond {
		t.Errorf("Default PolicyBuilder: expected InitialDelay=100ms, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 5*time.Second {
		t.Errorf("Default PolicyBuilder: expected MaxDelay=5s, got %v", policy.MaxDelay)
	}

	if policy.Multiplier != 2.0 {
		t.Errorf("Default PolicyBuilder: expected Multiplier=2.0, got %f", policy.Multiplier)
	}

	if !policy.Jitter {
		t.Error("Default PolicyBuilder: expected Jitter=true")
	}
}

// TestPolicyBuilder_FluentAPI tests the fluent API of PolicyBuilder
func TestPolicyBuilder_FluentAPI(t *testing.T) {
	t.Parallel()
	policy := NewPolicyBuilder().
		WithMaxRetries(10).
		WithInitialDelay(200 * time.Millisecond).
		WithMaxDelay(30 * time.Second).
		WithMultiplier(3.0).
		WithJitter(false).
		Build()

	if policy.MaxRetries != 10 {
		t.Errorf("FluentAPI: expected MaxRetries=10, got %d", policy.MaxRetries)
	}

	if policy.InitialDelay != 200*time.Millisecond {
		t.Errorf("FluentAPI: expected InitialDelay=200ms, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 30*time.Second {
		t.Errorf("FluentAPI: expected MaxDelay=30s, got %v", policy.MaxDelay)
	}

	if policy.Multiplier != 3.0 {
		t.Errorf("FluentAPI: expected Multiplier=3.0, got %f", policy.Multiplier)
	}

	if policy.Jitter {
		t.Error("FluentAPI: expected Jitter=false")
	}
}

// TestPolicyBuilder_PartialConfiguration tests partial configuration via fluent API
func TestPolicyBuilder_PartialConfiguration(t *testing.T) {
	t.Parallel()
	// Only set MaxRetries, others should be default
	policy := NewPolicyBuilder().
		WithMaxRetries(7).
		Build()

	if policy.MaxRetries != 7 {
		t.Errorf("Partial config: expected MaxRetries=7, got %d", policy.MaxRetries)
	}

	// Defaults should still be set
	if policy.InitialDelay != 100*time.Millisecond {
		t.Errorf("Partial config: expected default InitialDelay=100ms, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 5*time.Second {
		t.Errorf("Partial config: expected default MaxDelay=5s, got %v", policy.MaxDelay)
	}
}

// TestPolicyConfig_ToConfig tests converting Policy to PolicyConfig
func TestPolicyConfig_ToConfig(t *testing.T) {
	t.Parallel()
	policy := &Policy{
		MaxRetries:   5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     2 * time.Second,
		Multiplier:   2.5,
		Jitter:       true,
	}

	config := policy.ToConfig()

	if config.MaxRetries != 5 {
		t.Errorf("ToConfig: expected MaxRetries=5, got %d", config.MaxRetries)
	}

	if config.InitialDelayMs != 100 {
		t.Errorf("ToConfig: expected InitialDelayMs=100, got %d", config.InitialDelayMs)
	}

	if config.MaxDelayMs != 2000 {
		t.Errorf("ToConfig: expected MaxDelayMs=2000, got %d", config.MaxDelayMs)
	}

	if config.Multiplier != 2.5 {
		t.Errorf("ToConfig: expected Multiplier=2.5, got %f", config.Multiplier)
	}

	if !config.Jitter {
		t.Error("ToConfig: expected Jitter=true")
	}
}

// TestPolicyConfig_FromConfig tests converting PolicyConfig to Policy
func TestPolicyConfig_FromConfig(t *testing.T) {
	t.Parallel()
	config := PolicyConfig{
		MaxRetries:     7,
		InitialDelayMs: 250,
		MaxDelayMs:     10000,
		Multiplier:     3.0,
		Jitter:         false,
	}

	policy := FromConfig(config)

	if policy.MaxRetries != 7 {
		t.Errorf("FromConfig: expected MaxRetries=7, got %d", policy.MaxRetries)
	}

	if policy.InitialDelay != 250*time.Millisecond {
		t.Errorf("FromConfig: expected InitialDelay=250ms, got %v", policy.InitialDelay)
	}

	if policy.MaxDelay != 10*time.Second {
		t.Errorf("FromConfig: expected MaxDelay=10s, got %v", policy.MaxDelay)
	}

	if policy.Multiplier != 3.0 {
		t.Errorf("FromConfig: expected Multiplier=3.0, got %f", policy.Multiplier)
	}

	if policy.Jitter {
		t.Error("FromConfig: expected Jitter=false")
	}
}

// TestPolicyConfig_TOMLRoundtrip tests TOML serialization and deserialization
func TestPolicyConfig_TOMLRoundtrip(t *testing.T) {
	t.Parallel()
	original := &Policy{
		MaxRetries:   5,
		InitialDelay: 150 * time.Millisecond,
		MaxDelay:     3 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}

	// Convert to config
	config := original.ToConfig()

	// Serialize to TOML
	var tomlData string
	{
		var buf bytes.Buffer
		encoder := toml.NewEncoder(&buf)
		if err := encoder.Encode(config); err != nil {
			t.Fatalf("Failed to encode TOML: %v", err)
		}
		tomlData = buf.String()
	}

	// Deserialize from TOML
	var decoded PolicyConfig
	if _, err := toml.Decode(tomlData, &decoded); err != nil {
		t.Fatalf("Failed to decode TOML: %v", err)
	}

	// Convert back to Policy
	restored := FromConfig(decoded)

	// Compare values
	if restored.MaxRetries != original.MaxRetries {
		t.Errorf("Roundtrip: expected MaxRetries=%d, got %d", original.MaxRetries, restored.MaxRetries)
	}

	if restored.InitialDelay != original.InitialDelay {
		t.Errorf("Roundtrip: expected InitialDelay=%v, got %v", original.InitialDelay, restored.InitialDelay)
	}

	if restored.MaxDelay != original.MaxDelay {
		t.Errorf("Roundtrip: expected MaxDelay=%v, got %v", original.MaxDelay, restored.MaxDelay)
	}

	if restored.Multiplier != original.Multiplier {
		t.Errorf("Roundtrip: expected Multiplier=%f, got %f", original.Multiplier, restored.Multiplier)
	}

	if restored.Jitter != original.Jitter {
		t.Errorf("Roundtrip: expected Jitter=%v, got %v", original.Jitter, restored.Jitter)
	}
}

// TestPolicyConfig_TOMLFormat tests the actual TOML format
func TestPolicyConfig_TOMLFormat(t *testing.T) {
	t.Parallel()
	policy := DatabasePolicy()
	config := policy.ToConfig()

	// Encode to TOML
	var tomlData string
	{
		var buf bytes.Buffer
		encoder := toml.NewEncoder(&buf)
		if err := encoder.Encode(config); err != nil {
			t.Fatalf("Failed to encode TOML: %v", err)
		}
		tomlData = buf.String()
	}

	// Verify TOML contains expected keys
	expectedKeys := []string{
		"max_retries",
		"initial_delay_ms",
		"max_delay_ms",
		"multiplier",
		"jitter",
	}

	for _, key := range expectedKeys {
		if !contains(tomlData, key) {
			t.Errorf("TOML output missing expected key: %s", key)
		}
	}
}

// TestAllPredefinedPolicies tests that all predefined policies can be created
func TestAllPredefinedPolicies(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		policy *Policy
	}{
		{"DatabasePolicy", DatabasePolicy()},
		{"HTTPPolicy", HTTPPolicy()},
		{"FileIOPolicy", FileIOPolicy()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.policy == nil {
				t.Errorf("%s returned nil", tt.name)
			}

			if tt.policy.MaxRetries < 0 {
				t.Errorf("%s has invalid MaxRetries: %d", tt.name, tt.policy.MaxRetries)
			}

			if tt.policy.InitialDelay <= 0 {
				t.Errorf("%s has invalid InitialDelay: %v", tt.name, tt.policy.InitialDelay)
			}

			if tt.policy.MaxDelay <= 0 {
				t.Errorf("%s has invalid MaxDelay: %v", tt.name, tt.policy.MaxDelay)
			}

			if tt.policy.Multiplier <= 0 {
				t.Errorf("%s has invalid Multiplier: %f", tt.name, tt.policy.Multiplier)
			}
		})
	}
}

// TestPolicyBuilder_Chainability tests that all builder methods return the builder
func TestPolicyBuilder_Chainability(t *testing.T) {
	t.Parallel()
	builder := NewPolicyBuilder()

	// All methods should return the builder for chaining
	result := builder.
		WithMaxRetries(5).
		WithInitialDelay(50 * time.Millisecond).
		WithMaxDelay(2 * time.Second).
		WithMultiplier(2.0).
		WithJitter(true)

	if result != builder {
		t.Error("Builder methods should return the same builder instance for chaining")
	}
}

// TestPolicyConfig_EdgeCases tests edge cases in config conversion
func TestPolicyConfig_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		config PolicyConfig
	}{
		{
			name: "Zero values",
			config: PolicyConfig{
				MaxRetries:     0,
				InitialDelayMs: 0,
				MaxDelayMs:     0,
				Multiplier:     0,
				Jitter:         false,
			},
		},
		{
			name: "Large values",
			config: PolicyConfig{
				MaxRetries:     1000,
				InitialDelayMs: 60000,
				MaxDelayMs:     3600000,
				Multiplier:     10.0,
				Jitter:         true,
			},
		},
		{
			name: "Fractional multiplier",
			config: PolicyConfig{
				MaxRetries:     3,
				InitialDelayMs: 100,
				MaxDelayMs:     1000,
				Multiplier:     1.5,
				Jitter:         true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Convert to Policy and back
			policy := FromConfig(tt.config)
			roundtrip := policy.ToConfig()

			// Verify values match
			if roundtrip.MaxRetries != tt.config.MaxRetries {
				t.Errorf("MaxRetries mismatch: expected %d, got %d", tt.config.MaxRetries, roundtrip.MaxRetries)
			}

			if roundtrip.InitialDelayMs != tt.config.InitialDelayMs {
				t.Errorf("InitialDelayMs mismatch: expected %d, got %d", tt.config.InitialDelayMs, roundtrip.InitialDelayMs)
			}

			if roundtrip.MaxDelayMs != tt.config.MaxDelayMs {
				t.Errorf("MaxDelayMs mismatch: expected %d, got %d", tt.config.MaxDelayMs, roundtrip.MaxDelayMs)
			}

			if roundtrip.Multiplier != tt.config.Multiplier {
				t.Errorf("Multiplier mismatch: expected %f, got %f", tt.config.Multiplier, roundtrip.Multiplier)
			}

			if roundtrip.Jitter != tt.config.Jitter {
				t.Errorf("Jitter mismatch: expected %v, got %v", tt.config.Jitter, roundtrip.Jitter)
			}
		})
	}
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkDatabasePolicy benchmarks DatabasePolicy creation
func BenchmarkDatabasePolicy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DatabasePolicy()
	}
}

// BenchmarkHTTPPolicy benchmarks HTTPPolicy creation
func BenchmarkHTTPPolicy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = HTTPPolicy()
	}
}

// BenchmarkFileIOPolicy benchmarks FileIOPolicy creation
func BenchmarkFileIOPolicy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FileIOPolicy()
	}
}

// BenchmarkPolicyBuilder benchmarks PolicyBuilder usage
func BenchmarkPolicyBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewPolicyBuilder().
			WithMaxRetries(5).
			WithInitialDelay(100 * time.Millisecond).
			Build()
	}
}

// BenchmarkPolicyConfigConversion benchmarks Policy <-> Config conversion
func BenchmarkPolicyConfigConversion(b *testing.B) {
	policy := DatabasePolicy()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := policy.ToConfig()
		_ = FromConfig(config)
	}
}
