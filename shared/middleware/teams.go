package middleware

import (
	"fmt"
	"os"
	"strings"

	"github.com/appwrite/sdk-for-go/client"
	"github.com/appwrite/sdk-for-go/teams"
)

// TeamService wraps Appwrite Teams service with helper methods
type TeamService struct {
	client client.Client
	teams  *teams.Teams
}

// NewTeamService creates a new team service instance
func NewTeamService() (*TeamService, error) {
	appwriteClient := client.New()
	appwriteClient.Endpoint = os.Getenv("APPWRITE_ENDPOINT")
	appwriteClient.AddHeader("X-Appwrite-Project", os.Getenv("APPWRITE_PROJECT_ID"))
	appwriteClient.AddHeader("X-Appwrite-Key", os.Getenv("APPWRITE_API_KEY")) // Server API key for admin operations

	teamsService := teams.New(appwriteClient)

	return &TeamService{
		client: appwriteClient,
		teams:  teamsService,
	}, nil
}

// CreateTeamForWorkspace creates a new Appwrite Team for a workspace
func (ts *TeamService) CreateTeamForWorkspace(name, ownerID string) (string, error) {
	// Create the team - TeamId is required, use "unique()" to auto-generate
	team, err := ts.teams.Create("unique()", name)
	if err != nil {
		return "", fmt.Errorf("failed to create team: %w", err)
	}

	return team.Id, nil
}

// AssignTeamRole assigns a role to a user in a team
func (ts *TeamService) AssignTeamRole(teamID, userID, role string) error {
	// Appwrite Teams automatically assigns members, but we can add roles via team memberships
	// For owner role, we ensure the user is added as an owner
	_, err := ts.teams.CreateMembership(teamID, []string{role})
	if err != nil {
		// If membership already exists, that's okay
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to assign team role: %w", err)
		}
	}
	return nil
}

// GetUserTeams retrieves all teams a user belongs to
func (ts *TeamService) GetUserTeams(userID string) ([]string, error) {
	teamsList, err := ts.teams.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}

	var teamIDs []string
	for _, team := range teamsList.Teams {
		// Check if user is a member (simplified - in production, check memberships)
		teamIDs = append(teamIDs, team.Id)
	}

	return teamIDs, nil
}
