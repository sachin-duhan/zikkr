package git

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sachin-duhan/zikrr/pkg/util"
)

// RepositoryStatus represents the current status of a repository clone operation
type RepositoryStatus int

const (
	StatusPending RepositoryStatus = iota
	StatusCloning
	StatusRetrying
	StatusSuccess
	StatusFailed
	StatusSkipped
	StatusUpdating
)

func (s RepositoryStatus) String() string {
	switch s {
	case StatusPending:
		return "Pending"
	case StatusCloning:
		return "Cloning"
	case StatusRetrying:
		return "Retrying"
	case StatusSuccess:
		return "Success"
	case StatusFailed:
		return "Failed"
	case StatusSkipped:
		return "Skipped"
	case StatusUpdating:
		return "Updating"
	default:
		return "Unknown"
	}
}

// Repository represents a GitHub repository to be cloned
type Repository struct {
	Name         string
	Organization string
	URL          string
	Branch       string
	Status       RepositoryStatus
	Error        error
	Progress     string
	ExistingRepo ExistingRepoStrategy
	mu           sync.RWMutex
}

// RepositoryManager manages the state and operations of multiple repositories
type RepositoryManager struct {
	repositories []*Repository
	baseDir      string
	cloner       *ConcurrentCloner
	mu           sync.RWMutex
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(baseDir string, maxConcurrent int) *RepositoryManager {
	util.Info(fmt.Sprintf("Initializing repository manager with base directory: %s", baseDir))
	return &RepositoryManager{
		baseDir: baseDir,
		cloner:  NewConcurrentCloner(maxConcurrent),
	}
}

// AddRepository adds a new repository to be managed
func (rm *RepositoryManager) AddRepository(org, name, url, branch string, strategy ExistingRepoStrategy) *Repository {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	util.Debug(fmt.Sprintf("Adding repository to manager: %s/%s (branch: %s)", org, name, branch))
	repo := &Repository{
		Name:         name,
		Organization: org,
		URL:          url,
		Branch:       branch,
		Status:       StatusPending,
		ExistingRepo: strategy,
	}
	rm.repositories = append(rm.repositories, repo)
	return repo
}

// GetRepositories returns all managed repositories
func (rm *RepositoryManager) GetRepositories() []*Repository {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	repos := make([]*Repository, len(rm.repositories))
	copy(repos, rm.repositories)
	return repos
}

// CloneAll starts cloning all pending repositories
func (rm *RepositoryManager) CloneAll(ctx context.Context) <-chan *Repository {
	updates := make(chan *Repository, len(rm.repositories))
	util.Info(fmt.Sprintf("Starting clone of %d repositories", len(rm.repositories)))

	go func() {
		defer close(updates)

		// Prepare clone options for each repository
		cloneOpts := make([]CloneOptions, 0, len(rm.repositories))
		for _, repo := range rm.repositories {
			if repo.Status != StatusPending {
				util.Debug(fmt.Sprintf("Skipping non-pending repository: %s/%s (status: %s)", repo.Organization, repo.Name, repo.Status))
				continue
			}

			targetDir := filepath.Join(rm.baseDir, repo.Organization, repo.Name)
			util.Debug(fmt.Sprintf("Preparing to clone %s/%s to %s", repo.Organization, repo.Name, targetDir))

			opts := DefaultCloneOptions()
			opts.URL = repo.URL
			opts.TargetDir = targetDir
			opts.Branch = repo.Branch
			opts.ExistingRepo = repo.ExistingRepo
			opts.ProgressFunc = func(status string) {
				repo.mu.Lock()
				repo.Progress = status
				if strings.Contains(status, "Updating") {
					repo.Status = StatusUpdating
					util.Debug(fmt.Sprintf("Repository %s/%s is updating", repo.Organization, repo.Name))
				} else if strings.Contains(status, "Skipping") {
					repo.Status = StatusSkipped
					util.Debug(fmt.Sprintf("Repository %s/%s is skipped", repo.Organization, repo.Name))
				} else {
					repo.Status = StatusCloning
					util.Debug(fmt.Sprintf("Repository %s/%s is cloning", repo.Organization, repo.Name))
				}
				repo.mu.Unlock()
				updates <- repo
			}
			cloneOpts = append(cloneOpts, opts)
		}

		// Start cloning repositories
		results := rm.cloner.CloneRepositories(ctx, cloneOpts)
		for result := range results {
			// Find corresponding repository
			var repo *Repository
			for _, r := range rm.repositories {
				if r.URL == result.RepoURL {
					repo = r
					break
				}
			}
			if repo == nil {
				util.Error("Failed to find repository for result", fmt.Errorf("repository not found: %s", result.RepoURL))
				continue
			}

			// Update repository status
			repo.mu.Lock()
			if result.Success {
				if repo.Status != StatusSkipped {
					repo.Status = StatusSuccess
					util.Info(fmt.Sprintf("Repository %s/%s cloned successfully", repo.Organization, repo.Name))
				}
			} else {
				repo.Status = StatusFailed
				repo.Error = result.Error
				util.Error(fmt.Sprintf("Failed to clone repository %s/%s", repo.Organization, repo.Name), result.Error)
			}
			repo.mu.Unlock()
			updates <- repo
		}

		util.Info("Completed processing all repositories")
	}()

	return updates
}

// GetRepository returns a repository by its name and organization
func (rm *RepositoryManager) GetRepository(org, name string) *Repository {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for _, repo := range rm.repositories {
		if repo.Organization == org && repo.Name == name {
			return repo
		}
	}
	util.Debug(fmt.Sprintf("Repository not found: %s/%s", org, name))
	return nil
}

// UpdateStatus updates the status of a repository
func (r *Repository) UpdateStatus(status RepositoryStatus, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	oldStatus := r.Status
	r.Status = status
	if err != nil {
		r.Error = err
		util.Error(fmt.Sprintf("Repository %s/%s status changed from %s to %s with error", r.Organization, r.Name, oldStatus, status), err)
	} else {
		util.Debug(fmt.Sprintf("Repository %s/%s status changed from %s to %s", r.Organization, r.Name, oldStatus, status))
	}
}

// GetStatus returns the current status of a repository
func (r *Repository) GetStatus() (RepositoryStatus, error, string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.Status, r.Error, r.Progress
}

// SetExistingRepoStrategy sets the strategy for handling existing repositories
func (r *Repository) SetExistingRepoStrategy(strategy ExistingRepoStrategy) {
	r.mu.Lock()
	defer r.mu.Unlock()

	oldStrategy := r.ExistingRepo
	r.ExistingRepo = strategy
	util.Debug(fmt.Sprintf("Repository %s/%s strategy changed from %v to %v", r.Organization, r.Name, oldStrategy, strategy))
}
