package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
)

// RepositoryFilter defines criteria for filtering repositories
type RepositoryFilter struct {
	Visibility   string    // public, private, or all
	Topics       []string  // required topics
	UpdatedAfter time.Time // filter by last update time
	MinSize      int       // minimum size in KB
	MaxSize      int       // maximum size in KB
	Language     string    // primary language
	Archived     *bool     // filter archived repositories
	Fork         *bool     // filter forked repositories
}

// FilterRepositories filters a list of repositories based on the given criteria
func FilterRepositories(repos []*github.Repository, filter *RepositoryFilter) []*github.Repository {
	if filter == nil {
		return repos
	}

	filtered := make([]*github.Repository, 0, len(repos))
	for _, repo := range repos {
		if !matchesFilter(repo, filter) {
			continue
		}
		filtered = append(filtered, repo)
	}

	return filtered
}

// matchesFilter checks if a repository matches the filter criteria
func matchesFilter(repo *github.Repository, filter *RepositoryFilter) bool {
	// Check visibility
	if filter.Visibility != "" && filter.Visibility != "all" {
		isPrivate := repo.GetPrivate()
		if filter.Visibility == "public" && isPrivate {
			return false
		}
		if filter.Visibility == "private" && !isPrivate {
			return false
		}
	}

	// Check topics
	if len(filter.Topics) > 0 {
		repoTopics := repo.Topics
		if !hasAllTopics(repoTopics, filter.Topics) {
			return false
		}
	}

	// Check update time
	if !filter.UpdatedAfter.IsZero() {
		lastUpdate := repo.GetUpdatedAt().Time
		if lastUpdate.Before(filter.UpdatedAfter) {
			return false
		}
	}

	// Check size
	size := repo.GetSize()
	if filter.MinSize > 0 && size < filter.MinSize {
		return false
	}
	if filter.MaxSize > 0 && size > filter.MaxSize {
		return false
	}

	// Check language
	if filter.Language != "" {
		if !strings.EqualFold(repo.GetLanguage(), filter.Language) {
			return false
		}
	}

	// Check archived status
	if filter.Archived != nil {
		if repo.GetArchived() != *filter.Archived {
			return false
		}
	}

	// Check fork status
	if filter.Fork != nil {
		if repo.GetFork() != *filter.Fork {
			return false
		}
	}

	return true
}

// hasAllTopics checks if a repository has all required topics
func hasAllTopics(repoTopics []string, requiredTopics []string) bool {
	if len(requiredTopics) == 0 {
		return true
	}
	if len(repoTopics) == 0 {
		return false
	}

	topicMap := make(map[string]bool)
	for _, topic := range repoTopics {
		topicMap[strings.ToLower(topic)] = true
	}

	for _, required := range requiredTopics {
		if !topicMap[strings.ToLower(required)] {
			return false
		}
	}

	return true
}

// RepositoryInfo contains detailed information about a repository
type RepositoryInfo struct {
	*github.Repository
	Branches []*github.Branch
}

// GetRepositoryInfo fetches detailed information about a repository
func (c *Client) GetRepositoryInfo(ctx context.Context, owner, repo string) (*RepositoryInfo, error) {
	repository, err := c.GetRepository(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	branches, err := c.ListBranches(ctx, owner, repo, &github.BranchListOptions{
		Protected: nil,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	return &RepositoryInfo{
		Repository: repository,
		Branches:   branches,
	}, nil
}

// ListFilteredRepositories lists repositories in an organization with filtering
func (c *Client) ListFilteredRepositories(ctx context.Context, org string, filter *RepositoryFilter) ([]*github.Repository, error) {
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	// Set visibility in options if specified
	if filter != nil && filter.Visibility != "" && filter.Visibility != "all" {
		opts.Type = filter.Visibility
	}

	repos, err := c.ListOrganizationRepos(ctx, org, opts)
	if err != nil {
		return nil, err
	}

	return FilterRepositories(repos, filter), nil
}
