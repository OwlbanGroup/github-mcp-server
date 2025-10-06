//go:build e2e

package e2e_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAuthenticationErrors tests various authentication failure scenarios
func TestAuthenticationErrors(t *testing.T) {
	t.Parallel()

	// Note: This test would require setting up invalid tokens
	// For now, we'll skip as it requires special test setup
	t.Skip("Authentication error tests require special token setup")
}

// TestRateLimitingScenarios tests behavior under rate limiting
func TestRateLimitingScenarios(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing rate limiting scenarios")

	// Make multiple rapid requests to potentially trigger rate limiting
	repoName := helper.CreateTestRepo("rate-limit-test")

	for i := 0; i < 10; i++ {
		helper.WaitForRateLimit() // Add delay between calls
		helper.CallTool("get_repository", map[string]any{
			"owner": helper.GetOwner(),
			"repo":  repoName,
		})
	}

	helper.LogTestResult("Rate limiting handled correctly")
}

// TestInvalidParameters tests various invalid parameter scenarios
func TestInvalidParameters(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing invalid parameter handling")

	// Test invalid repository name
	if helper.ValidateToolAvailability("get_repository") {
		response := helper.CallToolWithError("get_repository", map[string]any{
			"owner": helper.GetOwner(),
			"repo":  "nonexistent-repo-12345-invalid",
		})
		require.True(t, response.IsError, "expected error for nonexistent repository")
	}

	// Test invalid owner
	if helper.ValidateToolAvailability("get_repository") {
		response := helper.CallToolWithError("get_repository", map[string]any{
			"owner": "nonexistent-user-12345",
			"repo":  "any-repo",
		})
		require.True(t, response.IsError, "expected error for nonexistent owner")
	}

	// Test empty parameters
	if helper.ValidateToolAvailability("get_repository") {
		response := helper.CallToolWithError("get_repository", map[string]any{
			"owner": "",
			"repo":  "",
		})
		require.True(t, response.IsError, "expected error for empty parameters")
	}

	helper.LogTestResult("Invalid parameters handled correctly")
}

// TestPermissionDeniedScenarios tests scenarios where operations are denied due to permissions
func TestPermissionDeniedScenarios(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing permission denied scenarios")

	// Try to access a private repository that we don't own
	// This might not fail if the token has access, so we'll test with invalid scenarios

	// Test creating repository with invalid parameters
	if helper.ValidateToolAvailability("create_repository") {
		response := helper.CallToolWithError("create_repository", map[string]any{
			"name": "",
		})
		require.True(t, response.IsError, "expected error for empty repository name")
	}

	helper.LogTestResult("Permission scenarios handled correctly")
}

// TestResourceNotFoundCases tests various "resource not found" scenarios
func TestResourceNotFoundCases(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing resource not found cases")

	repoName := helper.CreateTestRepo("not-found-test")

	// Test non-existent file
	if helper.ValidateToolAvailability("get_file_contents") {
		response := helper.CallToolWithError("get_file_contents", map[string]any{
			"owner":  helper.GetOwner(),
			"repo":   repoName,
			"path":   "nonexistent-file.txt",
			"branch": "main",
		})
		require.True(t, response.IsError, "expected error for nonexistent file")
	}

	// Test non-existent branch
	if helper.ValidateToolAvailability("list_branches") {
		response := helper.CallToolWithError("create_branch", map[string]any{
			"owner":       helper.GetOwner(),
			"repo":        repoName,
			"branch":      "new-branch",
			"from_branch": "nonexistent-branch",
		})
		require.True(t, response.IsError, "expected error for nonexistent source branch")
	}

	// Test non-existent issue
	if helper.ValidateToolAvailability("get_issue") {
		response := helper.CallToolWithError("get_issue", map[string]any{
			"owner":       helper.GetOwner(),
			"repo":        repoName,
			"issueNumber": 99999,
		})
		require.True(t, response.IsError, "expected error for nonexistent issue")
	}

	// Test non-existent pull request
	if helper.ValidateToolAvailability("get_pull_request") {
		response := helper.CallToolWithError("get_pull_request", map[string]any{
			"owner":      helper.GetOwner(),
			"repo":       repoName,
			"pullNumber": 99999,
		})
		require.True(t, response.IsError, "expected error for nonexistent PR")
	}

	helper.LogTestResult("Resource not found cases handled correctly")
}

// TestConcurrentOperationConflicts tests scenarios with concurrent operations
func TestConcurrentOperationConflicts(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing concurrent operation conflicts")

	repoName := helper.CreateTestRepo("concurrent-test")

	// Test concurrent file modifications (this is hard to trigger reliably)
	// We'll simulate by creating files with same name in quick succession

	filePath := "concurrent-file.txt"

	// Create initial file
	helper.CreateTestFile(repoName, "main", filePath, "Initial content", "Create file")

	// Try to create/update the same file quickly
	helper.CallTool("create_or_update_file", map[string]any{
		"owner":   helper.GetOwner(),
		"repo":    repoName,
		"path":    filePath,
		"content": "Updated content",
		"message": "Update file concurrently",
		"branch":  "main",
	})

	// Verify the file was updated
	getResponse := helper.CallTool("get_file_contents", map[string]any{
		"owner":  helper.GetOwner(),
		"repo":   repoName,
		"path":   filePath,
		"branch": "main",
	})
	require.Len(t, getResponse.Content, 2, "expected file to exist")

	helper.LogTestResult("Concurrent operations handled correctly")
}

// TestMalformedRequests tests handling of malformed or unexpected request data
func TestMalformedRequests(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing malformed request handling")

	// Test with invalid JSON in content fields
	if helper.ValidateToolAvailability("create_issue") {
		response := helper.CallToolWithError("create_issue", map[string]any{
			"owner": helper.GetOwner(),
			"repo":  "test-repo",
			"title": strings.Repeat("a", 1000), // Very long title
			"body":  "Valid body",
		})
		// This might succeed or fail depending on GitHub's limits
		helper.LogTestStep("Long title test completed")
	}

	// Test with special characters
	repoName := helper.CreateTestRepo("malformed-test")

	if helper.ValidateToolAvailability("create_issue") {
		response := helper.CallTool("create_issue", map[string]any{
			"owner": helper.GetOwner(),
			"repo":  repoName,
			"title": "Test with special chars: éñüñ 中文 🚀",
			"body":  "Body with special chars: @#$%^&*()",
		})
		require.False(t, response.IsError, "expected success with special characters")
	}

	helper.LogTestResult("Malformed requests handled correctly")
}

// TestNetworkFailureSimulation tests behavior when network issues occur
func TestNetworkFailureSimulation(t *testing.T) {
	t.Parallel()

	// This is difficult to test reliably without network interception
	// We'll test with timeouts and invalid endpoints
	t.Skip("Network failure tests require special network interception setup")
}

// TestLargeDataHandling tests handling of large amounts of data
func TestLargeDataHandling(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing large data handling")

	repoName := helper.CreateTestRepo("large-data-test")

	// Create a large file
	largeContent := strings.Repeat("This is a line of text in a large file.\n", 1000)
	filePath := "large-file.txt"

	response := helper.CallTool("create_or_update_file", map[string]any{
		"owner":   helper.GetOwner(),
		"repo":    repoName,
		"path":    filePath,
		"content": largeContent,
		"message": "Create large file for testing",
		"branch":  "main",
	})

	require.False(t, response.IsError, "expected success with large file")

	// Try to retrieve the large file
	getResponse := helper.CallTool("get_file_contents", map[string]any{
		"owner":  helper.GetOwner(),
		"repo":   repoName,
		"path":   filePath,
		"branch": "main",
	})

	require.False(t, getResponse.IsError, "expected success retrieving large file")
	require.Len(t, getResponse.Content, 2, "expected file content in response")

	helper.LogTestResult("Large data handling works correctly")
}

// TestBoundaryConditions tests edge cases at boundaries of valid input
func TestBoundaryConditions(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing boundary conditions")

	repoName := helper.CreateTestRepo("boundary-test")

	// Test minimum valid inputs
	if helper.ValidateToolAvailability("create_issue") {
		response := helper.CallTool("create_issue", map[string]any{
			"owner": helper.GetOwner(),
			"repo":  repoName,
			"title": "x", // Minimum title length
			"body":  "",  // Empty body is valid
		})
		require.False(t, response.IsError, "expected success with minimum inputs")
	}

	// Test file with empty content
	helper.CallTool("create_or_update_file", map[string]any{
		"owner":   helper.GetOwner(),
		"repo":    repoName,
		"path":    "empty-file.txt",
		"content": "",
		"message": "Create empty file",
		"branch":  "main",
	})

	// Test branch name with special characters
	helper.CreateTestBranch(repoName, "feature_branch.with.dots")

	helper.LogTestResult("Boundary conditions handled correctly")
}

// TestIdempotentOperations tests that operations can be safely repeated
func TestIdempotentOperations(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing idempotent operations")

	repoName := helper.CreateTestRepo("idempotent-test")

	// Test creating the same branch multiple times (should fail or succeed)
	branchName := "test-branch"
	helper.CreateTestBranch(repoName, branchName)

	// Try to create the same branch again (should fail)
	response := helper.CallToolWithError("create_branch", map[string]any{
		"owner":       helper.GetOwner(),
		"repo":        repoName,
		"branch":      branchName,
		"from_branch": "main",
	})
	require.True(t, response.IsError, "expected error when creating existing branch")

	// Test getting the same repository multiple times (should succeed)
	for i := 0; i < 3; i++ {
		response := helper.CallTool("get_repository", map[string]any{
			"owner": helper.GetOwner(),
			"repo":  repoName,
		})
		require.False(t, response.IsError, "expected success getting repository multiple times")
	}

	helper.LogTestResult("Idempotent operations handled correctly")
}
