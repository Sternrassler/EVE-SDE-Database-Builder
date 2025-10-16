package retry_test

import (
	"context"
	"fmt"
	"time"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/errors"
	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/retry"
)

// ExamplePolicy_Do demonstrates basic retry functionality
func ExamplePolicy_Do() {
	policy := retry.DefaultPolicy()
	ctx := context.Background()

	attempt := 0
	err := policy.Do(ctx, func() error {
		attempt++
		if attempt < 3 {
			return errors.NewRetryable("temporary failure", nil)
		}
		return nil // Success on 3rd attempt
	})

	if err == nil {
		fmt.Println("Operation succeeded after retries")
	}
	// Output: Operation succeeded after retries
}

// ExamplePolicy_Do_fatal demonstrates that fatal errors are not retried
func ExamplePolicy_Do_fatal() {
	policy := retry.DefaultPolicy()
	ctx := context.Background()

	attempt := 0
	err := policy.Do(ctx, func() error {
		attempt++
		return errors.NewFatal("permanent error", nil)
	})

	fmt.Printf("Attempts: %d\n", attempt)
	fmt.Printf("Error is fatal: %v\n", errors.IsFatal(err))
	// Output:
	// Attempts: 1
	// Error is fatal: true
}

// ExampleDoWithResult demonstrates retry with a return value
func ExampleDoWithResult() {
	policy := retry.DefaultPolicy()
	ctx := context.Background()

	attempt := 0
	result, err := retry.DoWithResult(ctx, policy, func() (string, error) {
		attempt++
		if attempt < 2 {
			return "", errors.NewRetryable("data not ready", nil)
		}
		return "success data", nil
	})

	if err == nil {
		fmt.Printf("Result: %s\n", result)
	}
	// Output: Result: success data
}

// ExampleNewPolicy demonstrates creating a custom retry policy
func ExampleNewPolicy() {
	// Create a policy with 5 retries, 50ms initial delay, 2s max delay
	policy := retry.NewPolicy(5, 50*time.Millisecond, 2*time.Second)

	ctx := context.Background()
	err := policy.Do(ctx, func() error {
		// Your operation here
		return nil
	})

	if err == nil {
		fmt.Println("Operation completed")
	}
	// Output: Operation completed
}

// ExampleDefaultPolicy demonstrates using the default policy
func ExampleDefaultPolicy() {
	policy := retry.DefaultPolicy()
	ctx := context.Background()

	err := policy.Do(ctx, func() error {
		// Simulated successful operation
		return nil
	})

	if err == nil {
		fmt.Println("Success with default policy")
	}
	// Output: Success with default policy
}

// ExampleDatabasePolicy demonstrates using the database-optimized policy
func ExampleDatabasePolicy() {
	policy := retry.DatabasePolicy()
	ctx := context.Background()

	err := policy.Do(ctx, func() error {
		// Simulated database operation
		return nil
	})

	if err == nil {
		fmt.Println("Database operation succeeded")
	}
	// Output: Database operation succeeded
}

// ExampleHTTPPolicy demonstrates using the HTTP-optimized policy
func ExampleHTTPPolicy() {
	policy := retry.HTTPPolicy()
	ctx := context.Background()

	err := policy.Do(ctx, func() error {
		// Simulated HTTP request
		return nil
	})

	if err == nil {
		fmt.Println("HTTP request succeeded")
	}
	// Output: HTTP request succeeded
}

// ExampleFileIOPolicy demonstrates using the file I/O policy
func ExampleFileIOPolicy() {
	policy := retry.FileIOPolicy()
	ctx := context.Background()

	err := policy.Do(ctx, func() error {
		// Simulated file operation
		return nil
	})

	if err == nil {
		fmt.Println("File operation succeeded")
	}
	// Output: File operation succeeded
}

// ExampleNewPolicyBuilder demonstrates using the policy builder
func ExampleNewPolicyBuilder() {
	policy := retry.NewPolicyBuilder().
		WithMaxRetries(5).
		WithInitialDelay(200 * time.Millisecond).
		WithMaxDelay(10 * time.Second).
		WithMultiplier(2.5).
		WithJitter(true).
		Build()

	ctx := context.Background()
	err := policy.Do(ctx, func() error {
		return nil
	})

	if err == nil {
		fmt.Println("Custom policy executed")
	}
	// Output: Custom policy executed
}

// ExampleFromConfig demonstrates creating a policy from configuration
func ExampleFromConfig() {
	cfg := retry.PolicyConfig{
		MaxRetries:     3,
		InitialDelayMs: 100,
		MaxDelayMs:     5000,
		Multiplier:     2.0,
		Jitter:         true,
	}

	policy := retry.FromConfig(cfg)
	ctx := context.Background()

	err := policy.Do(ctx, func() error {
		return nil
	})

	if err == nil {
		fmt.Println("Policy from config executed")
	}
	// Output: Policy from config executed
}

// ExamplePolicy_Do_contextCancellation demonstrates context cancellation
func ExamplePolicy_Do_contextCancellation() {
	policy := retry.NewPolicy(10, 100*time.Millisecond, 5*time.Second)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := policy.Do(ctx, func() error {
		return errors.NewRetryable("would retry", nil)
	})

	if err == context.Canceled {
		fmt.Println("Operation cancelled by context")
	}
	// Output: Operation cancelled by context
}
