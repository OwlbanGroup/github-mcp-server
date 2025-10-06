//go:build e2e

package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v74/github"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

// TestHelper provides common utilities for E2E tests
type TestHelper struct {
	t      *testing.T
	client *client.Client
	ctx    context.Context
	owner  string
}

// NewTestHelper creates a new test helper instance
func NewTestHelper(t *testing.T, client *client.Client) *TestHelper {
	ctx := context.Background()
	owner := getCurrentUser(t, client, ctx)

	return &TestHelper{
		t:      t,
		client: client,
		ctx:    ctx,
		owner:  owner,
	}
}

// getCurrentUser gets the current authenticated user
func getCurrentUser(t *testing.T, client *client.Client, ctx context.Context) string {
	request := mcp.CallToolRequest{}
	request.Params.Name = "get_me"

	response, err := client.CallTool(ctx, request)
	require.NoError(t, err, "expected to call 'get_me' tool successfully")
	require.False(t, response.IsError, "expected result not to be an error")

	var user struct {
		Login string `json:"login"`
	}
	err = json.Unmarshal([]byte(getTextContent(t, response)), &user)
	require.NoError(t, err, "expected to unmarshal user data")

	return user.Login
}

// getTextContent extracts text content from MCP response
func getTextContent(t *testing.T, response *mcp.CallToolResult) string {
	require.Len(t, response.Content, 1, "expected content to have one item")
	textContent, ok := response.Content[0].(mcp.TextContent)
	require.True(t, ok, "expected content to be of type TextContent")
	return textContent.Text
}

// CallTool calls a tool and returns the response
func (h *TestHelper) CallTool(toolName string, args map[string]any) *mcp.CallToolResult {
	request := mcp.CallToolRequest{}
	request.Params.Name = toolName
	request.Params.Arguments = args

	response, err := h.client.CallTool(h.ctx, request)
	require.NoError(h.t, err, "expected to call '%s' tool successfully", toolName)
	require.False(h.t, response.IsError, "expected '%s' result not to be an error", toolName)

	return response
}

// CallToolWithError calls a tool expecting an error
func (h *TestHelper) CallToolWithError(toolName string, args map[string]any) *mcp.CallToolResult {
	request := mcp.CallToolRequest{}
	request.Params.Name = toolName
	request.Params.Arguments = args

	response, err := h.client.CallTool(h.ctx, request)
	require.NoError(h.t, err, "expected to call '%s' tool successfully", toolName)
	require.True(h.t, response.IsError, "expected '%s' result to be an error", toolName)

	return response
}

// CreateTestRepo creates a temporary repository for testing
func (h *TestHelper) CreateTestRepo(name string) string {
	repoName := fmt.Sprintf("github-mcp-server-e2e-%s-%d", name, time.Now().UnixMilli())

	h.CallTool("create_repository", map[string]any{
		"name":     repoName,
		"private":  true,
		"autoInit": true,
	})

	// Register cleanup
	h.t.Cleanup(func() {
		h.DeleteTestRepo(repoName)
	})

	return repoName
}

// DeleteTestRepo deletes a test repository
func (h *TestHelper) DeleteTestRepo(repoName string) {
	ghClient := getRESTClient(h.t)
	_, err := ghClient.Repositories.Delete(context.Background(), h.owner, repoName)
	if err != nil {
		h.t.Logf("Warning: Failed to delete test repository %s/%s: %v", h.owner, repoName, err)
	}
}

// CreateTestBranch creates a test branch
func (h *TestHelper) CreateTestBranch(repoName, branchName string) {
	h.CallTool("create_branch", map[string]any{
		"owner":       h.owner,
		"repo":        repoName,
		"branch":      branchName,
		"from_branch": "main",
	})
}

// CreateTestFile creates a test file with content
func (h *TestHelper) CreateTestFile(repoName, branchName, filePath, content, message string) {
	h.CallTool("create_or_update_file", map[string]any{
		"owner":   h.owner,
		"repo":    repoName,
		"path":    filePath,
		"content": content,
		"message": message,
		"branch":  branchName,
	})
}

// CreateTestPR creates a test pull request
func (h *TestHelper) CreateTestPR(repoName, title, body, head, base string) int {
	response := h.CallTool("create_pull_request", map[string]any{
		"owner": h.owner,
		"repo":  repoName,
		"title": title,
		"body":  body,
		"head":  head,
		"base":  base,
	})

	var pr struct {
		Number int `json:"number"`
	}
	err := json.Unmarshal([]byte(getTextContent(h.t, response)), &pr)
	require.NoError(h.t, err, "expected to unmarshal PR data")

	return pr.Number
}

// CreateTestIssue creates a test issue
func (h *TestHelper) CreateTestIssue(repoName, title string) int {
	response := h.CallTool("create_issue", map[string]any{
		"owner": h.owner,
		"repo":  repoName,
		"title": title,
	})

	var issue struct {
		Number int `json:"number"`
	}
	err := json.Unmarshal([]byte(getTextContent(h.t, response)), &issue)
	require.NoError(h.t, err, "expected to unmarshal issue data")

	return issue.Number
}

// AssertJSONResponse asserts that the response contains expected JSON structure
func (h *TestHelper) AssertJSONResponse(response *mcp.CallToolResult, expected interface{}) {
	content := getTextContent(h.t, response)
	err := json.Unmarshal([]byte(content), expected)
	require.NoError(h.t, err, "expected to unmarshal JSON response")
}

// AssertTextResponse asserts that the response contains expected text
func (h *TestHelper) AssertTextResponse(response *mcp.CallToolResult, expected string) {
	content := getTextContent(h.t, response)
	require.Equal(h.t, expected, content, "expected text content to match")
}

// AssertContains asserts that the response contains a substring
func (h *TestHelper) AssertContains(response *mcp.CallToolResult, substring string) {
	content := getTextContent(h.t, response)
	require.Contains(h.t, content, substring, "expected content to contain substring")
}

// AssertArrayLength asserts that a JSON array response has expected length
func (h *TestHelper) AssertArrayLength(response *mcp.CallToolResult, expectedLength int) {
	content := getTextContent(h.t, response)
	var arr []interface{}
	err := json.Unmarshal([]byte(content), &arr)
	require.NoError(h.t, err, "expected to unmarshal array")
	require.Len(h.t, arr, expectedLength, "expected array to have correct length")
}

// GetOwner returns the current test owner
func (h *TestHelper) GetOwner() string {
	return h.owner
}

// LogTestStep logs a test step for better debugging
func (h *TestHelper) LogTestStep(step string) {
	h.t.Logf("🔄 %s", step)
}

// LogTestResult logs a test result
func (h *TestHelper) LogTestResult(result string) {
	h.t.Logf("✅ %s", result)
}

// GenerateUniqueName generates a unique name for test resources
func GenerateUniqueName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixMilli())
}

// WaitForRateLimit waits for rate limit reset if needed
func (h *TestHelper) WaitForRateLimit() {
	// Simple rate limit handling - wait 1 second between calls
	time.Sleep(1 * time.Second)
}

// ValidateToolAvailability checks if a tool is available in the current toolset
func (h *TestHelper) ValidateToolAvailability(toolName string) bool {
	request := mcp.ListToolsRequest{}
	response, err := h.client.ListTools(h.ctx, request)
	if err != nil {
		return false
	}

	for _, tool := range response.Tools {
		if tool.Name == toolName {
			return true
		}
	}
	return false
}

// SkipIfToolNotAvailable skips the test if the tool is not available
func (h *TestHelper) SkipIfToolNotAvailable(toolName string) {
	if !h.ValidateToolAvailability(toolName) {
		h.t.Skipf("Tool '%s' is not available in current toolset", toolName)
	}
}

// TestConfiguration holds test configuration
type TestConfiguration struct {
	Toolsets        []string
	Host            string
	ReadOnly        bool
	DynamicToolsets bool
}

// GetTestConfig returns the current test configuration
func GetTestConfig() TestConfiguration {
	config := TestConfiguration{
		Toolsets:        strings.Split(os.Getenv("GITHUB_TOOLSETS"), ","),
		Host:            getE2EHost(),
		ReadOnly:        os.Getenv("GITHUB_READ_ONLY") == "1",
		DynamicToolsets: os.Getenv("GITHUB_DYNAMIC_TOOLSETS") == "1",
	}

	if len(config.Toolsets) == 0 || (len(config.Toolsets) == 1 && config.Toolsets[0] == "") {
		config.Toolsets = github.GetDefaultToolsetIDs()
	}

	return config
}

// getE2EToken ensures the environment variable is checked only once and returns the token
func getE2EToken(t *testing.T) string {
	token := os.Getenv("GITHUB_MCP_SERVER_E2E_TOKEN")
	if token == "" {
		t.Fatalf("GITHUB_MCP_SERVER_E2E_TOKEN environment variable is not set")
	}
	return token
}

// getE2EHost ensures the environment variable is checked only once and returns the host
func getE2EHost() string {
	return os.Getenv("GITHUB_MCP_SERVER_E2E_HOST")
}

// getRESTClient creates a GitHub REST client for testing
func getRESTClient(t *testing.T) *github.Client {
	token := getE2EToken(t)
	ghClient := github.NewClient(nil).WithAuthToken(token)

	if host := getE2EHost(); host != "" && host != "https://github.com" {
		var err error
		ghClient, err = ghClient.WithEnterpriseURLs(host, host)
		require.NoError(t, err, "expected to create GitHub client with host")
	}

	return ghClient
}
