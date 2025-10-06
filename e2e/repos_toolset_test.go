//go:build e2e

package e2e_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestReposToolsetCreateRepository tests repository creation
func TestReposToolsetCreateRepository(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("create_repository")
	helper.LogTestStep("Testing repository creation")

	repoName := GenerateUniqueName("test-repo")

	response := helper.CallTool("create_repository", map[string]any{
		"name":        repoName,
		"description": "Test repository for E2E testing",
		"private":     true,
		"autoInit":    true,
	})

	var repo struct {
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		Description string `json:"description"`
		Private     bool   `json:"private"`
	}

	helper.AssertJSONResponse(response, &repo)
	require.Equal(t, repoName, repo.Name, "expected repository name to match")
	require.Contains(t, repo.FullName, helper.GetOwner(), "expected full name to contain owner")
	require.Equal(t, "Test repository for E2E testing", repo.Description, "expected description to match")
	require.True(t, repo.Private, "expected repository to be private")

	// Cleanup
	t.Cleanup(func() {
		helper.DeleteTestRepo(repoName)
	})

	helper.LogTestResult("Repository created successfully")
}

// TestReposToolsetGetRepository tests repository retrieval
func TestReposToolsetGetRepository(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("get_repository")
	helper.LogTestStep("Testing repository retrieval")

	repoName := helper.CreateTestRepo("get-repo-test")

	response := helper.CallTool("get_repository", map[string]any{
		"owner": helper.GetOwner(),
		"repo":  repoName,
	})

	var repo struct {
		Name     string `json:"name"`
		Owner    struct {
			Login string `json:"login"`
		} `json:"owner"`
		DefaultBranch string `json:"default_branch"`
	}

	helper.AssertJSONResponse(response, &repo)
	require.Equal(t, repoName, repo.Name, "expected repository name to match")
	require.Equal(t, helper.GetOwner(), repo.Owner.Login, "expected owner to match")
	require.Equal(t, "main", repo.DefaultBranch, "expected default branch to be main")

	helper.LogTestResult("Repository retrieved successfully")
}

// TestReposToolsetListRepositories tests repository listing
func TestReposToolsetListRepositories(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("list_repositories")
	helper.LogTestStep("Testing repository listing")

	response := helper.CallTool("list_repositories", map[string]any{
		"user": helper.GetOwner(),
	})

	var repos []struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	}

	helper.AssertJSONResponse(response, &repos)
	require.Greater(t, len(repos), 0, "expected at least one repository")

	// Verify that at least one repo belongs to the current user
	foundUserRepo := false
	for _, repo := range repos {
		if repo.Owner.Login == helper.GetOwner() {
			foundUserRepo = true
			break
		}
	}
	require.True(t, foundUserRepo, "expected to find at least one repository owned by the user")

	helper.LogTestResult("Repository listing works correctly")
}

// TestReposToolsetSearchRepositories tests repository search
func TestReposToolsetSearchRepositories(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("search_repositories")
	helper.LogTestStep("Testing repository search")

	// Create a test repo to search for
	repoName := helper.CreateTestRepo("search-test")

	response := helper.CallTool("search_repositories", map[string]any{
		"query": fmt.Sprintf("repo:%s/%s", helper.GetOwner(), repoName),
	})

	var searchResult struct {
		TotalCount int `json:"total_count"`
		Items      []struct {
			Name  string `json:"name"`
			Owner struct {
				Login string `json:"login"`
			} `json:"owner"`
		} `json:"items"`
	}

	helper.AssertJSONResponse(response, &searchResult)
	require.GreaterOrEqual(t, searchResult.TotalCount, 1, "expected at least one search result")
	require.GreaterOrEqual(t, len(searchResult.Items), 1, "expected at least one item in results")

	found := false
	for _, item := range searchResult.Items {
		if item.Name == repoName && item.Owner.Login == helper.GetOwner() {
			found = true
			break
		}
	}
	require.True(t, found, "expected to find the created repository in search results")

	helper.LogTestResult("Repository search works correctly")
}

// TestReposToolsetListBranches tests branch listing
func TestReposToolsetListBranches(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("list_branches")
	helper.LogTestStep("Testing branch listing")

	repoName := helper.CreateTestRepo("branches-test")

	response := helper.CallTool("list_branches", map[string]any{
		"owner": helper.GetOwner(),
		"repo":  repoName,
	})

	var branches []struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	helper.AssertJSONResponse(response, &branches)
	require.GreaterOrEqual(t, len(branches), 1, "expected at least one branch (main)")

	// Check that main branch exists
	mainBranchFound := false
	for _, branch := range branches {
		if branch.Name == "main" {
			mainBranchFound = true
			require.NotEmpty(t, branch.Commit.SHA, "expected main branch to have a commit SHA")
			break
		}
	}
	require.True(t, mainBranchFound, "expected main branch to exist")

	helper.LogTestResult("Branch listing works correctly")
}

// TestReposToolsetCreateBranch tests branch creation
func TestReposToolsetCreateBranch(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("create_branch")
	helper.LogTestStep("Testing branch creation")

	repoName := helper.CreateTestRepo("create-branch-test")
	branchName := "test-branch"

	response := helper.CallTool("create_branch", map[string]any{
		"owner":       helper.GetOwner(),
		"repo":        repoName,
		"branch":      branchName,
		"from_branch": "main",
	})

	var branch struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	helper.AssertJSONResponse(response, &branch)
	require.Equal(t, branchName, branch.Name, "expected branch name to match")
	require.NotEmpty(t, branch.Commit.SHA, "expected branch to have a commit SHA")

	helper.LogTestResult("Branch created successfully")
}

// TestReposToolsetFileOperations tests file creation, reading, and updating
func TestReposToolsetFileOperations(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("create_or_update_file")
	helper.SkipIfToolNotAvailable("get_file_contents")
	helper.LogTestStep("Testing file operations")

	repoName := helper.CreateTestRepo("file-ops-test")
	branchName := "main"
	filePath := "test-file.txt"
	initialContent := "Initial content for E2E test"
	updatedContent := "Updated content for E2E test"

	// Create file
	createResponse := helper.CallTool("create_or_update_file", map[string]any{
		"owner":   helper.GetOwner(),
		"repo":    repoName,
		"path":    filePath,
		"content": initialContent,
		"message": "Add test file",
		"branch":  branchName,
	})

	var createResult struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	helper.AssertJSONResponse(createResponse, &createResult)
	require.NotEmpty(t, createResult.Commit.SHA, "expected commit SHA")

	// Read file
	getResponse := helper.CallTool("get_file_contents", map[string]any{
		"owner":  helper.GetOwner(),
		"repo":   repoName,
		"path":   filePath,
		"branch": branchName,
	})

	// The response should contain embedded resource with file content
	require.Len(t, getResponse.Content, 2, "expected content and embedded resource")
	embeddedResource := getResponse.Content[1]
	textResource, ok := embeddedResource.(map[string]interface{})["resource"].(map[string]interface{})["textResourceContents"]
	require.True(t, ok, "expected embedded resource to contain text content")

	content := textResource.(map[string]interface{})["text"].(string)
	require.Equal(t, initialContent, content, "expected file content to match initial content")

	// Update file
	updateResponse := helper.CallTool("create_or_update_file", map[string]any{
		"owner":   helper.GetOwner(),
		"repo":    repoName,
		"path":    filePath,
		"content": updatedContent,
		"message": "Update test file",
		"branch":  branchName,
	})

	var updateResult struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	helper.AssertJSONResponse(updateResponse, &updateResult)
	require.NotEmpty(t, updateResult.Commit.SHA, "expected commit SHA")
	require.NotEqual(t, createResult.Commit.SHA, updateResult.Commit.SHA, "expected new commit SHA")

	helper.LogTestResult("File operations (create, read, update) work correctly")
}

// TestReposToolsetDeleteFile tests file deletion
func TestReposToolsetDeleteFile(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("delete_file")
	helper.LogTestStep("Testing file deletion")

	repoName := helper.CreateTestRepo("delete-file-test")
	filePath := "file-to-delete.txt"
	content := "This file will be deleted"

	// Create file first
	helper.CreateTestFile(repoName, "main", filePath, content, "Add file to delete")

	// Delete file
	deleteResponse := helper.CallTool("delete_file", map[string]any{
		"owner":   helper.GetOwner(),
		"repo":    repoName,
		"path":    filePath,
		"message": "Delete test file",
		"branch":  "main",
	})

	var deleteResult struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	helper.AssertJSONResponse(deleteResponse, &deleteResult)
	require.NotEmpty(t, deleteResult.Commit.SHA, "expected commit SHA")

	helper.LogTestResult("File deletion works correctly")
}

// TestReposToolsetCommits tests commit operations
func TestReposToolsetCommits(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("list_commits")
	helper.SkipIfToolNotAvailable("get_commit")
	helper.LogTestStep("Testing commit operations")

	repoName := helper.CreateTestRepo("commits-test")

	// List commits
	listResponse := helper.CallTool("list_commits", map[string]any{
		"owner": helper.GetOwner(),
		"repo":  repoName,
		"sha":   "main",
	})

	var commits []struct {
		SHA    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
		} `json:"commit"`
	}

	helper.AssertJSONResponse(listResponse, &commits)
	require.GreaterOrEqual(t, len(commits), 1, "expected at least one commit (initial)")

	latestCommit := commits[0]

	// Get specific commit
	getResponse := helper.CallTool("get_commit", map[string]any{
		"owner": helper.GetOwner(),
		"repo":  repoName,
		"sha":   latestCommit.SHA,
	})

	var commit struct {
		SHA    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
		} `json:"commit"`
		Files []struct {
			Filename string `json:"filename"`
		} `json:"files"`
	}

	helper.AssertJSONResponse(getResponse, &commit)
	require.Equal(t, latestCommit.SHA, commit.SHA, "expected commit SHA to match")
	require.Equal(t, latestCommit.Commit.Message, commit.Commit.Message, "expected commit message to match")

	helper.LogTestResult("Commit operations (list, get) work correctly")
}

// TestReposToolsetTags tests tag operations
func TestReposToolsetTags(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("list_tags")
	helper.SkipIfToolNotAvailable("get_tag")
	helper.LogTestStep("Testing tag operations")

	repoName := helper.CreateTestRepo("tags-test")

	// Create a tag using GitHub API (since MCP server may not support tag creation)
	ghClient := getRESTClient(t)
	ref, _, err := ghClient.Git.GetRef(t.Context(), helper.GetOwner(), repoName, "refs/heads/main")
	require.NoError(t, err, "expected to get main branch ref")

	tagObj, _, err := ghClient.Git.CreateTag(t.Context(), helper.GetOwner(), repoName, &gogithub.Tag{
		Tag:     gogithub.Ptr("v1.0.0"),
		Message: gogithub.Ptr("Test tag v1.0.0"),
		Object: &gogithub.GitObject{
			SHA:  ref.Object.SHA,
			Type: gogithub.Ptr("commit"),
		},
	})
	require.NoError(t, err, "expected to create tag object")

	_, _, err = ghClient.Git.CreateRef(t.Context(), helper.GetOwner(), repoName, &gogithub.Reference{
		Ref: gogithub.Ptr("refs/tags/v1.0.0"),
		Object: &gogithub.GitObject{
			SHA: tagObj.SHA,
		},
	})
	require.NoError(t, err, "expected to create tag ref")

	// List tags
	listResponse := helper.CallTool("list_tags", map[string]any{
		"owner": helper.GetOwner(),
		"repo":  repoName,
	})

	var tags []struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	helper.AssertJSONResponse(listResponse, &tags)
	require.GreaterOrEqual(t, len(tags), 1, "expected at least one tag")

	found := false
	for _, tag := range tags {
		if tag.Name == "v1.0.0" {
			found = true
			require.Equal(t, *ref.Object.SHA, tag.Commit.SHA, "expected tag commit SHA to match")
			break
		}
	}
	require.True(t, found, "expected to find v1.0.0 tag")

	// Get specific tag
	getResponse := helper.CallTool("get_tag", map[string]any{
		"owner": helper.GetOwner(),
		"repo":  repoName,
		"tag":   "v1.0.0",
	})

	var tag struct {
		Name   string `json:"name"`
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	helper.AssertJSONResponse(getResponse, &tag)
	require.Equal(t, "v1.0.0", tag.Name, "expected tag name to match")
	require.Equal(t, *ref.Object.SHA, tag.Commit.SHA, "expected tag commit SHA to match")

	helper.LogTestResult("Tag operations (list, get) work correctly")
}
