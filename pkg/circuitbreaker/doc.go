// Package circuitbreaker implements the circuit breaker resilience pattern.
//
// # Overview
//
// This package provides a circuit breaker to prevent cascading failures
// when calling external services or unreliable components:
//   - Three states: Closed, Open, Half-Open
//   - Configurable failure threshold and recovery timeout
//   - Fallback function support
//   - Thread-safe operations
//
// # Circuit Breaker States
//
//   - Closed: Normal operation, requests pass through
//   - Open: Circuit tripped, requests fail immediately
//   - Half-Open: Testing if service recovered
//
// # State Transitions
//
//	Closed → Open: After reaching failure threshold
//	Open → Half-Open: After recovery timeout
//	Half-Open → Closed: After successful request
//	Half-Open → Open: After failed request
//
// # Basic Usage
//
//	cb := circuitbreaker.New(
//	    circuitbreaker.WithFailureThreshold(5),
//	    circuitbreaker.WithRecoveryTimeout(10 * time.Second),
//	)
//
//	result, err := cb.Execute(func() (interface{}, error) {
//	    return callExternalService()
//	})
//
// # With Fallback
//
//	result, err := cb.ExecuteWithFallback(
//	    func() (interface{}, error) {
//	        return callPrimaryService()
//	    },
//	    func(err error) (interface{}, error) {
//	        return callBackupService()
//	    },
//	)
//
// # Configuration Options
//
//	cb := circuitbreaker.New(
//	    circuitbreaker.WithFailureThreshold(5),      // Open after 5 failures
//	    circuitbreaker.WithRecoveryTimeout(30*time.Second), // Wait 30s before retry
//	    circuitbreaker.WithSuccessThreshold(2),     // Need 2 successes to close
//	)
//
// # Monitoring
//
// Check circuit breaker state:
//
//	state := cb.State()
//	counts := cb.Counts()
//
// # Use Cases
//
//   - External API calls
//   - Database connections
//   - Microservice communication
//   - Third-party service integration
package circuitbreaker
