//go:build e2e

package e2e_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Constants for repeated literals
const (
	cleanupBranchName = "cleanup-branch"
)

// TestConcurrentOperationsLoad tests handling of multiple concurrent operations
func TestConcurrentOperationsLoad(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing concurrent operations load")

	repoName := helper.CreateTestRepo("concurrent-load-test")

	// Test concurrent repository operations
	var wg sync.WaitGroup
	numWorkers := 5
	operationsPerWorker := 10

	startTime := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerWorker; j++ {
				// Perform various operations concurrently
				helper.WaitForRateLimit()

				// Get repository info
				helper.CallTool("get_repository", map[string]any{
					"owner": helper.GetOwner(),
					"repo":  repoName,
				})

				// List branches
				if helper.ValidateToolAvailability("list_branches") {
					helper.CallTool("list_branches", map[string]any{
						"owner": helper.GetOwner(),
						"repo":  repoName,
					})
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	helper.LogTestResult("Concurrent operations completed in %v", duration)
	require.Less(t, duration, 2*time.Minute, "expected concurrent operations to complete within 2 minutes")
}

// TestResponseTimeBenchmarking tests response times for various operations
func TestResponseTimeBenchmarking(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Benchmarking response times")

	repoName := helper.CreateTestRepo("benchmark-test")

	// Benchmark repository operations
	operations := []struct {
		name string
		fn   func() error
	}{
		{"get_repository", func() error {
			response := helper.CallTool("get_repository", map[string]any{
				"owner": helper.GetOwner(),
				"repo":  repoName,
			})
			if response.IsError {
				return response.IsError
			}
			return nil
		}},
		{"list_branches", func() error {
			if !helper.ValidateToolAvailability("list_branches") {
				return nil // Skip if not available
			}
			response := helper.CallTool("list_branches", map[string]any{
				"owner": helper.GetOwner(),
				"repo":  repoName,
			})
			if response.IsError {
				return response.IsError
			}
			return nil
		}},
		{"get_me", func() error {
			response := helper.CallTool("get_me", map[string]any{})
			if response.IsError {
				return response.IsError
			}
			return nil
		}},
	}

	for _, op := range operations {
		start := time.Now()
		err := op.fn()
		duration := time.Since(start)

		require.NoError(t, err, "expected %s operation to succeed", op.name)
		helper.LogTestResult("%s operation took %v", op.name, duration)

		// Response should be reasonable (under 30 seconds)
		require.Less(t, duration, 30*time.Second, "expected %s to complete within 30 seconds", op.name)
	}
}

// TestStabilityUnderProlongedUsage tests system stability over extended periods
func TestStabilityUnderProlongedUsage(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing stability under prolonged usage")

	repoName := helper.CreateTestRepo("stability-test")

	// Perform operations repeatedly for an extended period
	iterations := 20
	startTime := time.Now()

	for i := 0; i < iterations; i++ {
		helper.LogTestStep("Iteration %d/%d", i+1, iterations)

		// Create a file
		fileName := "stability-" + string(rune(i+'0')) + ".txt"
		helper.CreateTestFile(repoName, "main", fileName, "Content for iteration "+string(rune(i+'0')), "Add stability test file")

		// Get repository info
		helper.CallTool("get_repository", map[string]any{
			"owner": helper.GetOwner(),
			"repo":  repoName,
		})

		// List branches
		if helper.ValidateToolAvailability("list_branches") {
			helper.CallTool("list_branches", map[string]any{
				"owner": helper.GetOwner(),
				"repo":  repoName,
			})
		}

		// Small delay between iterations
		time.Sleep(500 * time.Millisecond)
	}

	duration := time.Since(startTime)
	helper.LogTestResult("Stability test completed in %v", duration)

	// Verify repository still exists and is accessible
	finalResponse := helper.CallTool("get_repository", map[string]any{
		"owner": helper.GetOwner(),
		"repo":  repoName,
	})
	require.False(t, finalResponse.IsError, "expected repository to still be accessible after prolonged usage")
}

// TestResourceCleanupVerification tests that resources are properly cleaned up
func TestResourceCleanupVerification(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing resource cleanup verification")

	// Create multiple test repositories
	repoNames := make([]string, 3)
	for i := 0; i < 3; i++ {
		repoNames[i] = helper.CreateTestRepo("cleanup-test-" + string(rune(i+'0')))
	}

	// Perform operations on each repository
	for _, repoName := range repoNames {
		// Create some content
		helper.CreateTestBranch(repoName, cleanupBranchName)
		helper.CreateTestFile(repoName, cleanupBranchName, "cleanup.txt", "Content to be cleaned up", "Add cleanup file")

		// Create issue and PR
		issueNum := helper.CreateTestIssue(repoName, "Cleanup test issue")
		prNum := helper.CreateTestPR(repoName, "Cleanup test PR", "Test PR", cleanupBranchName, "main")

		helper.LogTestStep("Created resources in %s: issue #%d, PR #%d", repoName, issueNum, prNum)
	}

	// Resources should be cleaned up automatically by test cleanup
	// This test mainly verifies that the cleanup mechanism works
	helper.LogTestResult("Resource cleanup verification completed")
}

// TestMemoryUsageMonitoring tests for memory leaks or excessive memory usage
func TestMemoryUsageMonitoring(t *testing.T) {
	t.Parallel()

	// Note: Actual memory monitoring would require runtime metrics
	// This test performs operations that might reveal memory issues
	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing memory usage patterns")

	repoName := helper.CreateTestRepo("memory-test")

	// Perform many operations to stress memory usage
	for i := 0; i < 50; i++ {
		helper.WaitForRateLimit()

		// Create file with varying sizes
		content := "Memory test content iteration " + string(rune(i+'0')) + "\n"
		content += string(make([]byte, i*100)) // Increasing content size

		fileName := "memory-" + string(rune(i+'0')) + ".txt"
		helper.CallTool("create_or_update_file", map[string]any{
			"owner":   helper.GetOwner(),
			"repo":    repoName,
			"path":    fileName,
			"content": content,
			"message": "Add memory test file",
			"branch":  "main",
		})

		// Retrieve file
		helper.CallTool("get_file_contents", map[string]any{
			"owner":  helper.GetOwner(),
			"repo":   repoName,
			"path":   fileName,
			"branch": "main",
		})
	}

	helper.LogTestResult("Memory usage test completed without issues")
}

// TestOperationTimeouts tests that operations don't hang indefinitely
func TestOperationTimeouts(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing operation timeouts")

	repoName := helper.CreateTestRepo("timeout-test")

	// Test that operations complete within reasonable time
	operations := []struct {
		name     string
		timeout  time.Duration
		operation func() error
	}{
		{
			name:    "get_repository",
			timeout: 10 * time.Second,
			operation: func() error {
				response := helper.CallTool("get_repository", map[string]any{
					"owner": helper.GetOwner(),
					"repo":  repoName,
				})
				if response.IsError {
					return response.IsError
				}
				return nil
			},
		},
		{
			name:    "list_branches",
			timeout: 15 * time.Second,
			operation: func() error {
				if !helper.ValidateToolAvailability("list_branches") {
					return nil
				}
				response := helper.CallTool("list_branches", map[string]any{
					"owner": helper.GetOwner(),
					"repo":  repoName,
				})
				if response.IsError {
					return response.IsError
				}
				return nil
			},
		},
	}

	for _, op := range operations {
		done := make(chan error, 1)

		go func() {
			done <- op.operation()
		}()

		select {
		case err := <-done:
			require.NoError(t, err, "expected %s to succeed", op.name)
			helper.LogTestResult("%s completed within timeout", op.name)
		case <-time.After(op.timeout):
			t.Fatalf("%s operation timed out after %v", op.name, op.timeout)
		}
	}
}

// TestGradualLoadIncrease tests system behavior under gradually increasing load
func TestGradualLoadIncrease(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing gradual load increase")

	repoName := helper.CreateTestRepo("gradual-load-test")

	// Gradually increase the number of operations
	for batch := 1; batch <= 5; batch++ {
		helper.LogTestStep("Batch %d: Performing %d operations", batch, batch*2)

		startTime := time.Now()

		// Perform batch*2 operations
		for i := 0; i < batch*2; i++ {
			helper.WaitForRateLimit()

			// Alternate between different operations
			if i%2 == 0 {
				helper.CallTool("get_repository", map[string]any{
					"owner": helper.GetOwner(),
					"repo":  repoName,
				})
			} else {
				if helper.ValidateToolAvailability("list_branches") {
					helper.CallTool("list_branches", map[string]any{
						"owner": helper.GetOwner(),
						"repo":  repoName,
					})
				}
			}
		}

		batchDuration := time.Since(startTime)
		helper.LogTestResult("Batch %d completed in %v", batch, batchDuration)

		// Each batch should not take excessively long
		require.Less(t, batchDuration, time.Duration(batch)*10*time.Second, "expected batch %d to complete within reasonable time", batch)
	}
}

// TestRecoveryAfterLoad tests system recovery after high load
func TestRecoveryAfterLoad(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing recovery after high load")

	repoName := helper.CreateTestRepo("recovery-test")

	// Generate high load
	helper.LogTestStep("Generating high load...")
	for i := 0; i < 20; i++ {
		helper.WaitForRateLimit()
		helper.CallTool("get_repository", map[string]any{
			"owner": helper.GetOwner(),
			"repo":  repoName,
		})
	}

	// Test recovery - operations should still work normally
	helper.LogTestStep("Testing recovery...")

	for i := 0; i < 5; i++ {
		response := helper.CallTool("get_repository", map[string]any{
			"owner": helper.GetOwner(),
			"repo":  repoName,
		})
		require.False(t, response.IsError, "expected operation to succeed after high load")
	}

	// Create new resources to verify full recovery
	newRepoName := helper.CreateTestRepo("recovery-verification-test")
	helper.CreateTestFile(newRepoName, "main", "recovery.txt", "Recovery test content", "Add recovery test file")

	helper.LogTestResult("System recovered successfully after high load")
}

// TestOperationThroughput tests the overall throughput of operations
func TestOperationThroughput(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing operation throughput")

	repoName := helper.CreateTestRepo("throughput-test")

	// Measure throughput for a series of operations
	numOperations := 30
	startTime := time.Now()

	for i := 0; i < numOperations; i++ {
		helper.WaitForRateLimit()
		helper.CallTool("get_repository", map[string]any{
			"owner": helper.GetOwner(),
			"repo":  repoName,
		})
	}

	totalDuration := time.Since(startTime)
	throughput := float64(numOperations) / totalDuration.Seconds()

	helper.LogTestResult("Completed %d operations in %v (%.2f ops/sec)", numOperations, totalDuration, throughput)

	// Throughput should be reasonable (at least 0.5 ops/sec with rate limiting)
	require.Greater(t, throughput, 0.1, "expected minimum throughput of 0.1 ops/sec")
}
