package auth

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// TokenType represents the type of GitHub token
type TokenType int

const (
	// TokenTypeClassic represents a classic GitHub PAT
	TokenTypeClassic TokenType = iota
	// TokenTypeFineGrained represents a fine-grained GitHub PAT
	TokenTypeFineGrained
)

// Token represents a GitHub authentication token with its metadata
type Token struct {
	Value     string
	Type      TokenType
	ExpiresAt *github.Timestamp
	Client    *github.Client
}

// ValidateToken validates the GitHub token and returns its metadata
func ValidateToken(ctx context.Context, tokenValue string) (*Token, error) {
	log.Printf("[DEBUG] Validating GitHub token")
	if tokenValue == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	// Create GitHub client
	client := github.NewClient(nil).WithAuthToken(tokenValue)

	// Get authenticated user to validate token
	user, resp, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Printf("[ERROR] Failed to validate token: %v", err)
		if resp != nil && resp.StatusCode == 401 {
			return nil, fmt.Errorf("invalid token: authentication failed")
		}
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	log.Printf("[INFO] Token validated successfully for user: %s", *user.Login)

	// Determine token type based on response headers
	tokenType := TokenTypeClassic
	if resp.Header.Get("GitHub-Authentication-Token-Type") == "fine-grained" {
		tokenType = TokenTypeFineGrained
		log.Printf("[DEBUG] Detected fine-grained token")
	} else {
		log.Printf("[DEBUG] Detected classic token")
	}

	return &Token{
		Value:  tokenValue,
		Type:   tokenType,
		Client: client,
	}, nil
}

// CheckOrganizationAccess verifies if the token has access to the specified organization
func (t *Token) CheckOrganizationAccess(ctx context.Context, orgName string) (bool, error) {
	log.Printf("[DEBUG] Checking access to organization: %s", orgName)

	org, resp, err := t.Client.Organizations.Get(ctx, orgName)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			log.Printf("[DEBUG] No access to organization %s or it doesn't exist", orgName)
			return false, nil
		}
		return false, fmt.Errorf("error checking organization access: %w", err)
	}

	if org != nil {
		log.Printf("[INFO] Token has access to organization: %s", *org.Login)
		return true, nil
	}

	return false, nil
}

// CheckRepositoryAccess checks if the token has access to a specific repository in an organization
func (t *Token) CheckRepositoryAccess(ctx context.Context, orgName, repoName string) (bool, error) {
	log.Printf("[DEBUG] Checking access to repository: %s/%s", orgName, repoName)

	repo, resp, err := t.Client.Repositories.Get(ctx, orgName, repoName)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			log.Printf("[DEBUG] No access to repository %s/%s or it doesn't exist", orgName, repoName)
			return false, nil
		}
		return false, fmt.Errorf("error checking repository access: %w", err)
	}

	if repo != nil {
		log.Printf("[INFO] Token has access to repository: %s/%s", orgName, *repo.Name)
		return true, nil
	}

	return false, nil
}

// GetTokenFromEnv attempts to get a GitHub token from environment variables
func GetTokenFromEnv() string {
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		log.Printf("[DEBUG] Found token in GITHUB_TOKEN env var: %s****", token[:4])
		return token
	}

	token = os.Getenv("ZIKRR_GITHUB_TOKEN")
	if token != "" {
		log.Printf("[DEBUG] Found token in ZIKRR_GITHUB_TOKEN env var: %s****", token[:4])
	}
	return token
}

// CreateGitHubClient creates a new GitHub client with the given token
func CreateGitHubClient(ctx context.Context, token *Token) *github.Client {
	log.Printf("[DEBUG] Creating GitHub client with token")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token.Value},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}
