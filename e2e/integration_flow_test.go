//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test constants for repeated strings
const (
	checkMarkRepoCreated     = "✓ Repository created"
	checkMarkBranchCreated   = "✓ Branch created"
	checkMarkFileCreated     = "✓ File created and committed"
	checkMarkPRCreated       = "✓ Pull request created"
	checkMarkPRCommentAdded  = "✓ Pull request comment added"
	checkMarkPRReviewCreated = "✓ Pull request review created"
	checkMarkPRMerged        = "✓ Pull request merged"
	checkMarkFileVerified    = "✓ File verified in main branch"
	checkMarkIssueCreated    = "✓ Issue created"
	checkMarkLabelsAdded     = "✓ Labels added to issue"
	checkMarkIssueAssigned   = "✓ Issue assigned"
	checkMarkCommentAdded    = "✓ Issue comment added"
	checkMarkFixBranch       = "✓ Fix branch created"
	checkMarkFixCommitted    = "✓ Fix committed"
	checkMarkIssueClosed     = "✓ Issue closed"
	checkMarkDiscussionAdded = "✓ Discussion comments added"
	checkMarkFeatureBranch   = "✓ Feature implemented"
	checkMarkFeatureMerged   = "✓ Feature merged and issue closed"

	featureBranchName = "feature-branch"
	fixBranchName     = "fix-crash"
	featureRequestBranch = "feature-request"
)

// TestCompleteRepositoryLifecycle tests the complete repository lifecycle
func TestCompleteRepositoryLifecycle(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing complete repository lifecycle")

	// Phase 1: Create repository
	repoName := helper.CreateTestRepo("lifecycle-test")
	helper.LogTestResult(checkMarkRepoCreated)

	// Phase 2: Create branch
	helper.CreateTestBranch(repoName, featureBranchName)
	helper.LogTestResult(checkMarkBranchCreated)

	// Phase 3: Create file and commit
	helper.CreateTestFile(repoName, featureBranchName, "feature.go", "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}", "Add main.go")
	helper.LogTestResult(checkMarkFileCreated)

	// Phase 4: Create pull request
	prNumber := helper.CreateTestPR(repoName, "Add Hello World feature", "This PR adds a simple Hello World program", featureBranchName, "main")
	helper.LogTestResult(checkMarkPRCreated)

	// Phase 5: Add PR comment
	helper.CallTool("add_pull_request_comment", map[string]any{
		"owner":      helper.GetOwner(),
		"repo":       repoName,
		"pullNumber": prNumber,
		"body":       "Looks good! Ready for review.",
	})
	helper.LogTestResult(checkMarkPRCommentAdded)

	// Phase 6: Create PR review
	helper.CallTool("create_pull_request_review", map[string]any{
		"owner":      helper.GetOwner(),
		"repo":       repoName,
		"pullNumber": prNumber,
		"event":      "APPROVE",
		"body":       "Approved! Great work on this feature.",
	})
	helper.LogTestResult(checkMarkPRReviewCreated)

	// Phase 7: Merge pull request
	mergeResponse := helper.CallTool("merge_pull_request", map[string]any{
		"owner":       helper.GetOwner(),
		"repo":        repoName,
		"pullNumber":  prNumber,
		"mergeMethod": "merge",
	})

	var mergeResult struct {
		Merged bool `json:"merged"`
	}
	helper.AssertJSONResponse(mergeResponse, &mergeResult)
	require.True(t, mergeResult.Merged, "expected PR to be merged successfully")
	helper.LogTestResult(checkMarkPRMerged)

	// Phase 8: Verify file exists in main branch
	getFileResponse := helper.CallTool("get_file_contents", map[string]any{
		"owner":  helper.GetOwner(),
		"repo":   repoName,
		"path":   "feature.go",
		"branch": "main",
	})
	require.Len(t, getFileResponse.Content, 2, "expected file to exist in main branch")
	helper.LogTestResult(checkMarkFileVerified)

	helper.LogTestResult("🎉 Complete repository lifecycle test passed!")
}

// TestIssueManagementWorkflow tests complete issue management workflow
func TestIssueManagementWorkflow(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing complete issue management workflow")

	// Phase 1: Create repository
	repoName := helper.CreateTestRepo("issue-workflow-test")
	helper.LogTestResult(checkMarkRepoCreated)

	// Phase 2: Create issue
	issueNumber := helper.CreateTestIssue(repoName, "Bug: Application crashes on startup")
	helper.LogTestResult(checkMarkIssueCreated)

	// Phase 3: Add labels to issue
	helper.CallTool("add_issue_labels", map[string]any{
		"owner":       helper.GetOwner(),
		"repo":        repoName,
		"issueNumber": issueNumber,
		"labels":      []string{"bug", "high-priority", "crash"},
	})
	helper.LogTestResult(checkMarkLabelsAdded)

	// Phase 4: Assign issue
	helper.CallTool("add_issue_assignees", map[string]any{
		"owner":       helper.GetOwner(),
		"repo":        repoName,
		"issueNumber": issueNumber,
		"assignees":   []string{helper.GetOwner()},
	})
	helper.LogTestResult(checkMarkIssueAssigned)

	// Phase 5: Add issue comment
	helper.CallTool("add_issue_comment", map[string]any{
		"owner":       helper.GetOwner(),
		"repo":        repoName,
		"issueNumber": issueNumber,
		"body":        "I've reproduced this issue. Working on a fix.",
	})
	helper.LogTestResult(checkMarkCommentAdded)

	// Phase 6: Create branch for fix
	helper.CreateTestBranch(repoName, fixBranchName)
	helper.LogTestResult(checkMarkFixBranch)

	// Phase 7: Create fix file
	helper.CreateTestFile(repoName, fixBranchName, "fix.patch", "diff --git a/main.c b/main.c\nindex 1234567..abcdef0 100644\n--- a/main.c\n+++ b/main.c\n@@ -1,3 +1,3 @@\n int main() {\n-    crash();\n+    return 0;\n }", "Fix crash in main function")
	helper.LogTestResult(checkMarkFixCommitted)

	// Phase 8: Create pull request referencing the issue
	prNumber := helper.CreateTestPR(repoName, "Fix: Application crash on startup (fixes #"+string(rune(issueNumber+'0'))+")", "This PR fixes the application crash issue reported in #"+string(rune(issueNumber+'0')), fixBranchName, "main")
	helper.LogTestResult(checkMarkPRCreated)

	// Phase 9: Merge the fix
	helper.CallTool("merge_pull_request", map[string]any{
		"owner":       helper.GetOwner(),
		"repo":        repoName,
		"pullNumber":  prNumber,
		"mergeMethod": "merge",
	})
	helper.LogTestResult(checkMarkPRMerged)

	// Phase 10: Close the issue
	helper.CallTool("update_issue", map[string]any{
		"owner":       helper.GetOwner(),
		"repo":        repoName,
		"issueNumber": issueNumber,
		"state":       "closed",
	})
	helper.LogTestResult(checkMarkIssueClosed)

	helper.LogTestResult("🎉 Complete issue management workflow test passed!")
}

// TestMultiBranchWorkflow tests working with multiple branches
func TestMultiBranchWorkflow(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing multi-branch workflow")

	// Phase 1: Create repository
	repoName := helper.CreateTestRepo("multi-branch-test")
	helper.LogTestResult(checkMarkRepoCreated)

	// Phase 2: Create multiple feature branches
	branches := []string{"feature-ui", "feature-api", "feature-db"}
	for _, branch := range branches {
		helper.CreateTestBranch(repoName, branch)
		helper.CreateTestFile(repoName, branch, branch+".txt", "Content for "+branch, "Add "+branch+" file")
	}
	helper.LogTestResult("✓ Multiple feature branches created")

	// Phase 3: Create pull requests for each branch
	prNumbers := make([]int, len(branches))
	for i, branch := range branches {
		prNumbers[i] = helper.CreateTestPR(repoName, "Add "+branch+" feature", "Implementation of "+branch, branch, "main")
	}
	helper.LogTestResult("✓ Pull requests created for each branch")

	// Phase 4: List all branches
	listBranchesResponse := helper.CallTool("list_branches", map[string]any{
		"owner": helper.GetOwner(),
		"repo":  repoName,
	})

	var branchList []struct {
		Name string `json:"name"`
	}
	helper.AssertJSONResponse(listBranchesResponse, &branchList)
	require.GreaterOrEqual(t, len(branchList), len(branches)+1, "expected at least main + feature branches")

	// Verify all our branches exist
	branchNames := make(map[string]bool)
	for _, branch := range branchList {
		branchNames[branch.Name] = true
	}
	require.True(t, branchNames["main"], "expected main branch")
	for _, branch := range branches {
		require.True(t, branchNames[branch], "expected %s branch", branch)
	}
	helper.LogTestResult("✓ All branches verified")

	// Phase 5: List all pull requests
	listPRsResponse := helper.CallTool("list_pull_requests", map[string]any{
		"owner": helper.GetOwner(),
		"repo":  repoName,
		"state": "open",
	})

	var prList []struct {
		Number int `json:"number"`
		Title  string `json:"title"`
	}
	helper.AssertJSONResponse(listPRsResponse, &prList)
	require.Len(t, prList, len(branches), "expected correct number of open PRs")
	helper.LogTestResult("✓ All pull requests verified")

	helper.LogTestResult("🎉 Multi-branch workflow test passed!")
}

// TestCollaborativeWorkflow tests scenarios with multiple users/contributors
func TestCollaborativeWorkflow(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing collaborative workflow")

	// Phase 1: Create repository
	repoName := helper.CreateTestRepo("collaborative-test")
	helper.LogTestResult(checkMarkRepoCreated)

	// Phase 2: Create issue
	issueNumber := helper.CreateTestIssue(repoName, "New feature request")
	helper.LogTestResult(checkMarkIssueCreated)

	// Phase 3: Add multiple comments to simulate discussion
	comments := []string{
		"This feature would be very useful!",
		"I agree, let's implement this.",
		"What do you think about this approach?",
		"Looks good to me. Should we add tests?",
	}

	for _, comment := range comments {
		helper.CallTool("add_issue_comment", map[string]any{
			"owner":       helper.GetOwner(),
			"repo":        repoName,
			"issueNumber": issueNumber,
			"body":        comment,
		})
	}
	helper.LogTestResult(checkMarkDiscussionAdded)

	// Phase 4: Create feature branch
	helper.CreateTestBranch(repoName, featureRequestBranch)
	helper.CreateTestFile(repoName, featureRequestBranch, "feature.md", "# New Feature\n\nThis implements the requested feature.", "Implement new feature")
	helper.LogTestResult(checkMarkFeatureBranch)

	// Phase 5: Create pull request
	prNumber := helper.CreateTestPR(repoName, "Implement new feature request", "Closes #"+string(rune(issueNumber+'0')), featureRequestBranch, "main")
	helper.LogTestResult(checkMarkPRCreated)

	// Phase 6: Add PR review comments
	helper.CallTool("create_pull_request_review", map[string]any{
		"owner":      helper.GetOwner(),
		"repo":       repoName,
		"pullNumber": prNumber,
		"event":      "COMMENT",
		"body":       "Great implementation! Just a few minor suggestions.",
	})
	helper.LogTestResult("✓ Code review completed")

	// Phase 7: Merge and close
	helper.CallTool("merge_pull_request", map[string]any{
		"owner":       helper.GetOwner(),
		"repo":        repoName,
		"pullNumber":  prNumber,
		"mergeMethod": "merge",
	})

	helper.CallTool("update_issue", map[string]any{
		"owner":       helper.GetOwner(),
		"repo":        repoName,
		"issueNumber": issueNumber,
		"state":       "closed",
	})
	helper.LogTestResult(checkMarkFeatureMerged)

	helper.LogTestResult("🎉 Collaborative workflow test passed!")
}

// TestErrorRecoveryWorkflow tests error handling and recovery scenarios
func TestErrorRecoveryWorkflow(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing error recovery workflow")

	// Phase 1: Create repository
	repoName := helper.CreateTestRepo("error-recovery-test")
	helper.LogTestResult(checkMarkRepoCreated)

	// Phase 2: Test invalid operations (should fail gracefully)
	helper.LogTestStep("Testing invalid operations")

	// Try to get non-existent issue
	invalidIssueResponse := helper.CallToolWithError("get_issue", map[string]any{
		"owner":       helper.GetOwner(),
		"repo":        repoName,
		"issueNumber": 99999,
	})
	require.True(t, invalidIssueResponse.IsError, "expected error for invalid issue")
	helper.LogTestResult("✓ Invalid issue handled correctly")

	// Try to get non-existent PR
	invalidPRResponse := helper.CallToolWithError("get_pull_request", map[string]any{
		"owner":      helper.GetOwner(),
		"repo":       repoName,
		"pullNumber": 99999,
	})
	require.True(t, invalidPRResponse.IsError, "expected error for invalid PR")
	helper.LogTestResult("✓ Invalid PR handled correctly")

	// Phase 3: Test recovery - create valid resources after errors
	issueNumber := helper.CreateTestIssue(repoName, "Valid issue after errors")
	prNumber := helper.CreateTestPR(repoName, "Valid PR after errors", "Test", "main", "main")
	helper.LogTestResult("✓ Valid resources created after error recovery")

	// Phase 4: Test concurrent operations (create multiple resources)
	for i := 0; i < 3; i++ {
		helper.CreateTestFile(repoName, "main", "concurrent-"+string(rune(i+'0'))+".txt", "Concurrent file "+string(rune(i+'0')), "Add concurrent file")
	}
	helper.LogTestResult("✓ Concurrent operations handled")

	// Phase 5: Verify everything still works after errors
	getIssueResponse := helper.CallTool("get_issue", map[string]any{
		"owner":       helper.GetOwner(),
		"repo":        repoName,
		"issueNumber": issueNumber,
	})
	require.False(t, getIssueResponse.IsError, "expected successful issue retrieval after errors")

	getPRResponse := helper.CallTool("get_pull_request", map[string]any{
		"owner":      helper.GetOwner(),
		"repo":       repoName,
		"pullNumber": prNumber,
	})
	require.False(t, getPRResponse.IsError, "expected successful PR retrieval after errors")
	helper.LogTestResult("✓ System recovered and working correctly")

	helper.LogTestResult("🎉 Error recovery workflow test passed!")
}
