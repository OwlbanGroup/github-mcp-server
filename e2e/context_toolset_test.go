//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestContextToolsetGetMe tests the get_me tool from context toolset
func TestContextToolsetGetMe(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing get_me tool")

	// Test successful get_me call
	response := helper.CallTool("get_me", map[string]any{})

	var user struct {
		Login string `json:"login"`
		ID    int    `json:"id"`
		Type  string `json:"type"`
	}

	helper.AssertJSONResponse(response, &user)
	require.NotEmpty(t, user.Login, "expected user login to be non-empty")
	require.Greater(t, user.ID, 0, "expected user ID to be greater than 0")
	require.NotEmpty(t, user.Type, "expected user type to be non-empty")

	helper.LogTestResult("get_me tool returned valid user information")
}

// TestContextToolsetGetTeams tests the get_teams tool from context toolset
func TestContextToolsetGetTeams(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("get_teams")
	helper.LogTestStep("Testing get_teams tool")

	// Test get_teams without user parameter (gets authenticated user's teams)
	response := helper.CallTool("get_teams", map[string]any{})

	var teams []struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	helper.AssertJSONResponse(response, &teams)
	// Note: teams array might be empty if user has no teams, which is valid

	helper.LogTestResult("get_teams tool returned team list")
}

// TestContextToolsetGetTeamsWithUser tests the get_teams tool with user parameter
func TestContextToolsetGetTeamsWithUser(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("get_teams")
	helper.LogTestStep("Testing get_teams tool with user parameter")

	// Test get_teams with authenticated user
	response := helper.CallTool("get_teams", map[string]any{
		"user": helper.GetOwner(),
	})

	var teams []struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	helper.AssertJSONResponse(response, &teams)

	helper.LogTestResult("get_teams tool with user parameter returned team list")
}

// TestContextToolsetGetTeamMembers tests the get_team_members tool from context toolset
func TestContextToolsetGetTeamMembers(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.SkipIfToolNotAvailable("get_team_members")
	helper.LogTestStep("Testing get_team_members tool")

	// First, get user's teams to find a team to test with
	teamsResponse := helper.CallTool("get_teams", map[string]any{})

	var teams []struct {
		Name       string `json:"name"`
		Slug       string `json:"slug"`
		Organization struct {
			Login string `json:"login"`
		} `json:"organization"`
	}

	helper.AssertJSONResponse(teamsResponse, &teams)

	if len(teams) == 0 {
		t.Skip("No teams available for testing get_team_members")
	}

	// Test get_team_members with first available team
	team := teams[0]
	response := helper.CallTool("get_team_members", map[string]any{
		"org":      team.Organization.Login,
		"team_slug": team.Slug,
	})

	var members []struct {
		Login string `json:"login"`
		ID    int    `json:"id"`
		Type  string `json:"type"`
	}

	helper.AssertJSONResponse(response, &members)
	require.GreaterOrEqual(t, len(members), 1, "expected at least one team member (the authenticated user)")

	helper.LogTestResult("get_team_members tool returned team member list")
}

// TestContextToolsetInvalidParameters tests error handling for invalid parameters
func TestContextToolsetInvalidParameters(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Testing invalid parameters for context tools")

	// Test get_team_members with invalid org
	if helper.ValidateToolAvailability("get_team_members") {
		response := helper.CallToolWithError("get_team_members", map[string]any{
			"org":       "nonexistent-org-12345",
			"team_slug": "nonexistent-team",
		})
		require.True(t, response.IsError, "expected error for nonexistent org/team")
	}

	helper.LogTestResult("Context tools properly handle invalid parameters")
}

// TestContextToolsetToolAvailability tests that context tools are available
func TestContextToolsetToolAvailability(t *testing.T) {
	t.Parallel()

	mcpClient := setupMCPClient(t)
	helper := NewTestHelper(t, mcpClient)

	helper.LogTestStep("Verifying context toolset availability")

	// Context tools should always be available (strongly recommended)
	require.True(t, helper.ValidateToolAvailability("get_me"), "get_me tool should be available")

	// Other context tools may or may not be available depending on configuration
	helper.LogTestResult("Context toolset availability verified")
}
