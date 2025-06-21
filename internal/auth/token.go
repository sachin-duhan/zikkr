package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	Scopes    []string
	ExpiresAt *github.Timestamp
}

// ValidateToken validates the GitHub token and returns its metadata
func ValidateToken(ctx context.Context, tokenValue string) (*Token, error) {
	if tokenValue == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	// Create OAuth2 client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tokenValue},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Get token metadata
	metadata, _, err := client.Auth.GetTokenMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token metadata: %w", err)
	}

	// Determine token type
	tokenType := TokenTypeClassic
	if metadata.ExpiresAt != nil {
		tokenType = TokenTypeFineGrained
	}

	// Get token scopes
	scopes := []string{}
	if metadata.Scopes != nil {
		scopes = *metadata.Scopes
	}

	token := &Token{
		Value:     tokenValue,
		Type:      tokenType,
		Scopes:    scopes,
		ExpiresAt: metadata.ExpiresAt,
	}

	// Validate required scopes
	if err := validateRequiredScopes(token); err != nil {
		return nil, err
	}

	return token, nil
}

// validateRequiredScopes checks if the token has the required scopes
func validateRequiredScopes(token *Token) error {
	requiredScopes := map[string]bool{
		"repo":     false,
		"read:org": false,
	}

	// For classic tokens, check the exact scope names
	if token.Type == TokenTypeClassic {
		for _, scope := range token.Scopes {
			if _, ok := requiredScopes[scope]; ok {
				requiredScopes[scope] = true
			}
		}
	} else {
		// For fine-grained tokens, check for repository and organization permissions
		for _, scope := range token.Scopes {
			if strings.HasPrefix(scope, "repository:") {
				requiredScopes["repo"] = true
			}
			if strings.HasPrefix(scope, "organization:") {
				requiredScopes["read:org"] = true
			}
		}
	}

	// Check if any required scopes are missing
	missingScopes := []string{}
	for scope, found := range requiredScopes {
		if !found {
			missingScopes = append(missingScopes, scope)
		}
	}

	if len(missingScopes) > 0 {
		return fmt.Errorf("token missing required scopes: %s", strings.Join(missingScopes, ", "))
	}

	return nil
}

// GetTokenFromEnv attempts to get a GitHub token from environment variables
func GetTokenFromEnv() string {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("ZIKRR_GITHUB_TOKEN")
	}
	return token
}

// CreateGitHubClient creates a new GitHub client with the given token
func CreateGitHubClient(ctx context.Context, token *Token) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token.Value},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}
