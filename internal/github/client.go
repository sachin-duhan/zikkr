package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/sachin-duhan/zikrr/internal/auth"
	"github.com/sachin-duhan/zikrr/pkg/util"
)

// Client wraps the GitHub client with additional functionality
type Client struct {
	client *github.Client
	token  *auth.Token
}

// RateLimitInfo contains information about the current rate limit status
type RateLimitInfo struct {
	Remaining int
	Limit     int
	Reset     time.Time
}

// NewClient creates a new GitHub client with the given token
func NewClient(ctx context.Context, token *auth.Token) *Client {
	return &Client{
		client: auth.CreateGitHubClient(ctx, token),
		token:  token,
	}
}

// GetRateLimit returns the current rate limit status
func (c *Client) GetRateLimit(ctx context.Context) (*RateLimitInfo, error) {
	limits, _, err := c.client.RateLimits(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limits: %w", err)
	}

	core := limits.Core
	return &RateLimitInfo{
		Remaining: core.Remaining,
		Limit:     core.Limit,
		Reset:     core.Reset.Time,
	}, nil
}

// WaitForRateLimit waits until the rate limit resets if necessary
func (c *Client) WaitForRateLimit(ctx context.Context) error {
	info, err := c.GetRateLimit(ctx)
	if err != nil {
		return err
	}

	if info.Remaining > 0 {
		return nil
	}

	waitDuration := time.Until(info.Reset)
	if waitDuration <= 0 {
		return nil
	}

	util.Info(fmt.Sprintf("Rate limit exceeded. Waiting %v for reset...", waitDuration.Round(time.Second)))

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitDuration):
		return nil
	}
}

// GetOrganization gets information about a GitHub organization
func (c *Client) GetOrganization(ctx context.Context, name string) (*github.Organization, error) {
	if err := c.WaitForRateLimit(ctx); err != nil {
		return nil, err
	}

	org, _, err := c.client.Organizations.Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization %q: %w", name, err)
	}

	return org, nil
}

// ListOrganizationRepos lists all repositories in an organization with pagination
func (c *Client) ListOrganizationRepos(ctx context.Context, org string, opts *github.RepositoryListByOrgOptions) ([]*github.Repository, error) {
	if err := c.WaitForRateLimit(ctx); err != nil {
		return nil, err
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := c.client.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for organization %q: %w", org, err)
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// GetRepository gets information about a specific repository
func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*github.Repository, error) {
	if err := c.WaitForRateLimit(ctx); err != nil {
		return nil, err
	}

	repository, _, err := c.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository %s/%s: %w", owner, repo, err)
	}

	return repository, nil
}

// ListBranches lists all branches in a repository
func (c *Client) ListBranches(ctx context.Context, owner, repo string, opts *github.BranchListOptions) ([]*github.Branch, error) {
	if err := c.WaitForRateLimit(ctx); err != nil {
		return nil, err
	}

	var allBranches []*github.Branch
	for {
		branches, resp, err := c.client.Repositories.ListBranches(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list branches for repository %s/%s: %w", owner, repo, err)
		}

		allBranches = append(allBranches, branches...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allBranches, nil
}
