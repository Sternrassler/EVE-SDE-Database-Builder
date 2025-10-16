// Package retry provides exponential backoff retry logic for transient errors.
package retry

import (
	"time"
)

// DatabasePolicy returns a retry policy optimized for database operations
// 5 retries, 50ms initial delay, 2s max delay
func DatabasePolicy() *Policy {
	return &Policy{
		MaxRetries:   5,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     2 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// HTTPPolicy returns a retry policy optimized for HTTP requests
// 3 retries, 100ms initial delay, 5s max delay, jitter enabled
func HTTPPolicy() *Policy {
	return &Policy{
		MaxRetries:   3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// FileIOPolicy returns a retry policy optimized for file I/O operations
// 2 retries, 10ms initial delay, 500ms max delay
func FileIOPolicy() *Policy {
	return &Policy{
		MaxRetries:   2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// PolicyBuilder provides a fluent API for constructing custom retry policies
type PolicyBuilder struct {
	policy *Policy
}

// NewPolicyBuilder creates a new PolicyBuilder with default values
func NewPolicyBuilder() *PolicyBuilder {
	return &PolicyBuilder{
		policy: &Policy{
			MaxRetries:   3,
			InitialDelay: 100 * time.Millisecond,
			MaxDelay:     5 * time.Second,
			Multiplier:   2.0,
			Jitter:       true,
		},
	}
}

// WithMaxRetries sets the maximum number of retry attempts
func (pb *PolicyBuilder) WithMaxRetries(n int) *PolicyBuilder {
	pb.policy.MaxRetries = n
	return pb
}

// WithInitialDelay sets the initial delay before the first retry
func (pb *PolicyBuilder) WithInitialDelay(d time.Duration) *PolicyBuilder {
	pb.policy.InitialDelay = d
	return pb
}

// WithMaxDelay sets the maximum delay between retries
func (pb *PolicyBuilder) WithMaxDelay(d time.Duration) *PolicyBuilder {
	pb.policy.MaxDelay = d
	return pb
}

// WithMultiplier sets the exponential backoff multiplier
func (pb *PolicyBuilder) WithMultiplier(m float64) *PolicyBuilder {
	pb.policy.Multiplier = m
	return pb
}

// WithJitter enables or disables jitter
func (pb *PolicyBuilder) WithJitter(enabled bool) *PolicyBuilder {
	pb.policy.Jitter = enabled
	return pb
}

// Build returns the constructed Policy
func (pb *PolicyBuilder) Build() *Policy {
	return pb.policy
}

// PolicyConfig represents a policy configuration for TOML serialization
type PolicyConfig struct {
	MaxRetries     int     `toml:"max_retries"`
	InitialDelayMs int64   `toml:"initial_delay_ms"`
	MaxDelayMs     int64   `toml:"max_delay_ms"`
	Multiplier     float64 `toml:"multiplier"`
	Jitter         bool    `toml:"jitter"`
}

// ToConfig converts a Policy to a PolicyConfig for TOML serialization
func (p *Policy) ToConfig() PolicyConfig {
	return PolicyConfig{
		MaxRetries:     p.MaxRetries,
		InitialDelayMs: p.InitialDelay.Milliseconds(),
		MaxDelayMs:     p.MaxDelay.Milliseconds(),
		Multiplier:     p.Multiplier,
		Jitter:         p.Jitter,
	}
}

// FromConfig creates a Policy from a PolicyConfig (TOML deserialization)
func FromConfig(cfg PolicyConfig) *Policy {
	return &Policy{
		MaxRetries:   cfg.MaxRetries,
		InitialDelay: time.Duration(cfg.InitialDelayMs) * time.Millisecond,
		MaxDelay:     time.Duration(cfg.MaxDelayMs) * time.Millisecond,
		Multiplier:   cfg.Multiplier,
		Jitter:       cfg.Jitter,
	}
}
