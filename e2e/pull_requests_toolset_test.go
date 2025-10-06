//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Constants for repeated literals
const (
	featureBranchName = "feature-branch"
	mergeBranchName   = "merge-branch"

	expectedTitleMatch = "expected title to match"
)

// TestPullRequestsToolsetCreatePullRequest tests PR creation
func TestPullRequestsToolsetCreatePullRequest(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("create_pull_request")
	helper.LogTestStep("Testing pull request creation")

	repoName := helper.CreateTestRepo("pr-create-test")
	helper.CreateTestBranch(repoName, featureBranchName)
	helper.CreateTestFile(repoName, featureBranchName, "feature.txt", "New feature content", "Add feature file")

	response := helper.CallTool("create_pull_request", map[string]any{
		"owner": helper.GetOwner(),
		"repo":  repoName,
		"title": "Test Pull Request",
		"body":  "This is a test PR created by E2E tests.",
		"head":  featureBranchName,
		"base":  "main",
	})

	var pr struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		State  string `json:"state"`
		Head   struct {
			Ref string `json:"ref"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
	}

	helper.AssertJSONResponse(response, &pr)
	require.Equal(t, 1, pr.Number, "expected PR number to be 1")
	require.Equal(t, "Test Pull Request", pr.Title, expectedTitleMatch)
	require.Equal(t, "This is a test PR created by E2E tests.", pr.Body, "expected body to match")
	require.Equal(t, "open", pr.State, "expected PR to be open")
	require.Equal(t, "feature-branch", pr.Head.Ref, "expected head branch to match")
	require.Equal(t, "main", pr.Base.Ref, "expected base branch to match")

	helper.LogTestResult("Pull request created successfully")
}

// TestPullRequestsToolsetGetPullRequest tests PR retrieval
func TestPullRequestsToolsetGetPullRequest(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("get_pull_request")
	helper.LogTestStep("Testing pull request retrieval")

	repoName := helper.CreateTestRepo("pr-get-test")
	prNumber := helper.CreateTestPR(repoName, "PR to Retrieve", "Test PR body", "main", "main")

	response := helper.CallTool("get_pull_request", map[string]any{
		"owner":      helper.GetOwner(),
		"repo":       repoName,
		"pullNumber": prNumber,
	})

	var pr struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		State  string `json:"state"`
	}

	helper.AssertJSONResponse(response, &pr)
	require.Equal(t, prNumber, pr.Number, "expected PR number to match")
	require.Equal(t, "PR to Retrieve", pr.Title, expectedTitleMatch)
	require.Equal(t, "open", pr.State, "expected PR to be open")

	helper.LogTestResult("Pull request retrieved successfully")
}

// TestPullRequestsToolsetListPullRequests tests PR listing
func TestPullRequestsToolsetListPullRequests(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("list_pull_requests")
	helper.LogTestStep("Testing pull request listing")

	repoName := helper.CreateTestRepo("pr-list-test")

	// Create multiple PRs
	pr1 := helper.CreateTestPR(repoName, "First PR", "First PR body", "main", "main")
	pr2 := helper.CreateTestPR(repoName, "Second PR", "Second PR body", "main", "main")

	response := helper.CallTool("list_pull_requests", map[string]any{
		"owner": helper.GetOwner(),
		"repo":  repoName,
		"state": "open",
	})

	var prs []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		State  string `json:"state"`
	}

	helper.AssertJSONResponse(response, &prs)
	require.GreaterOrEqual(t, len(prs), 2, "expected at least 2 PRs")

	// Verify our test PRs are in the list
	foundPR1 := false
	foundPR2 := false
	for _, pr := range prs {
		if pr.Number == pr1 {
			foundPR1 = true
			require.Equal(t, "First PR", pr.Title, expectedTitleMatch)
		}
		if pr.Number == pr2 {
			foundPR2 = true
			require.Equal(t, "Second PR", pr.Title, expectedTitleMatch)
		}
	}
	require.True(t, foundPR1, "expected to find first PR")
	require.True(t, foundPR2, "expected to find second PR")

	helper.LogTestResult("Pull request listing works correctly")
}

// TestPullRequestsToolsetUpdatePullRequest tests PR updating
func TestPullRequestsToolsetUpdatePullRequest(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("update_pull_request")
	helper.LogTestStep("Testing pull request updating")

	repoName := helper.CreateTestRepo("pr-update-test")
	prNumber := helper.CreateTestPR(repoName, "PR to Update", "Original body", "main", "main")

	response := helper.CallTool("update_pull_request", map[string]any{
		"owner":      helper.GetOwner(),
		"repo":       repoName,
		"pullNumber": prNumber,
		"title":      "Updated PR Title",
		"body":       "Updated PR body with new content.",
		"state":      "closed",
	})

	var pr struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		State  string `json:"state"`
	}

	helper.AssertJSONResponse(response, &pr)
	require.Equal(t, prNumber, pr.Number, "expected PR number to match")
	require.Equal(t, "Updated PR Title", pr.Title, "expected title to be updated")
	require.Equal(t, "Updated PR body with new content.", pr.Body, "expected body to be updated")
	require.Equal(t, "closed", pr.State, "expected PR to be closed")

	helper.LogTestResult("Pull request updated successfully")
}

// TestPullRequestsToolsetCreatePullRequestReview tests PR review creation
func TestPullRequestsToolsetCreatePullRequestReview(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("create_pull_request_review")
	helper.LogTestStep("Testing pull request review creation")

	repoName := helper.CreateTestRepo("pr-review-test")
	prNumber := helper.CreateTestPR(repoName, "PR for Review", "Test PR", "main", "main")

	response := helper.CallTool("create_pull_request_review", map[string]any{
		"owner":      helper.GetOwner(),
		"repo":       repoName,
		"pullNumber": prNumber,
		"event":      "COMMENT",
		"body":       "This is a test review comment.",
	})

	var review struct {
		ID     int    `json:"id"`
		State  string `json:"state"`
		Body   string `json:"body"`
		User   struct {
			Login string `json:"login"`
		} `json:"user"`
	}

	helper.AssertJSONResponse(response, &review)
	require.Greater(t, review.ID, 0, "expected review ID to be greater than 0")
	require.Equal(t, "COMMENTED", review.State, "expected review state to be COMMENTED")
	require.Equal(t, "This is a test review comment.", review.Body, "expected review body to match")
	require.Equal(t, helper.GetOwner(), review.User.Login, "expected review user to match")

	helper.LogTestResult("Pull request review created successfully")
}

// TestPullRequestsToolsetGetPullRequestReviews tests PR review retrieval
func TestPullRequestsToolsetGetPullRequestReviews(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("get_pull_request_reviews")
	helper.LogTestStep("Testing pull request review retrieval")

	repoName := helper.CreateTestRepo("pr-reviews-get-test")
	prNumber := helper.CreateTestPR(repoName, "PR for Reviews", "Test PR", "main", "main")

	// Create a review first
	helper.CallTool("create_pull_request_review", map[string]any{
		"owner":      helper.GetOwner(),
		"repo":       repoName,
		"pullNumber": prNumber,
		"event":      "COMMENT",
		"body":       "Test review for retrieval",
	})

	response := helper.CallTool("get_pull_request_reviews", map[string]any{
		"owner":      helper.GetOwner(),
		"repo":       repoName,
		"pullNumber": prNumber,
	})

	var reviews []struct {
		State string `json:"state"`
		Body  string `json:"body"`
	}

	helper.AssertJSONResponse(response, &reviews)
	require.GreaterOrEqual(t, len(reviews), 1, "expected at least one review")

	found := false
	for _, review := range reviews {
		if review.Body == "Test review for retrieval" {
			found = true
			require.Equal(t, "COMMENTED", review.State, "expected review state to match")
			break
		}
	}
	require.True(t, found, "expected to find the test review")

	helper.LogTestResult("Pull request reviews retrieved successfully")
}

// TestPullRequestsToolsetMergePullRequest tests PR merging
func TestPullRequestsToolsetMergePullRequest(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("merge_pull_request")
	helper.LogTestStep("Testing pull request merging")

	repoName := helper.CreateTestRepo("pr-merge-test")

	// Create a branch with changes
	helper.CreateTestBranch(repoName, mergeBranchName)
	helper.CreateTestFile(repoName, mergeBranchName, "merge.txt", "Content to merge", "Add file for merge")

	prNumber := helper.CreateTestPR(repoName, "PR to Merge", "Test PR for merging", mergeBranchName, "main")

	response := helper.CallTool("merge_pull_request", map[string]any{
		"owner":                helper.GetOwner(),
		"repo":                 repoName,
		"pullNumber":           prNumber,
		"mergeMethod":          "merge",
		"commitTitle":          "Merge PR: Test PR for merging",
		"commitMessage":        "Merging test PR via E2E tests",
	})

	var mergeResult struct {
		Merged  bool   `json:"merged"`
		Message string `json:"message"`
		SHA     string `json:"sha"`
	}

	helper.AssertJSONResponse(response, &mergeResult)
	require.True(t, mergeResult.Merged, "expected PR to be merged")
	require.NotEmpty(t, mergeResult.SHA, "expected merge commit SHA")

	helper.LogTestResult("Pull request merged successfully")
}

// TestPullRequestsToolsetInvalidPullRequestNumber tests error handling for invalid PR numbers
func TestPullRequestsToolsetInvalidPullRequestNumber(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("get_pull_request")
	helper.LogTestStep("Testing invalid pull request number handling")

	repoName := helper.CreateTestRepo("invalid-pr-test")

	// Try to get a non-existent PR
	response := helper.CallToolWithError("get_pull_request", map[string]any{
		"owner":      helper.GetOwner(),
		"repo":       repoName,
		"pullNumber": 99999,
	})

	require.True(t, response.IsError, "expected error for non-existent PR")

	helper.LogTestResult("Invalid pull request numbers handled correctly")
}
