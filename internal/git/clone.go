package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/sachin-duhan/zikrr/pkg/util"
)

// ExistingRepoStrategy defines how to handle existing repositories
type ExistingRepoStrategy int

const (
	// SkipExisting skips cloning if the repository already exists
	SkipExisting ExistingRepoStrategy = iota
	// OverwriteExisting deletes and re-clones if the repository exists
	OverwriteExisting
	// FetchOnly updates the existing repository without re-cloning
	FetchOnly
)

// CloneOptions represents options for cloning a repository
type CloneOptions struct {
	URL          string
	TargetDir    string
	Branch       string
	Timeout      time.Duration
	MaxRetries   int
	ProgressFunc func(status string)
	ConnTimeout  time.Duration
	CloneTimeout time.Duration
	ExistingRepo ExistingRepoStrategy
}

// DefaultCloneOptions returns default clone options
func DefaultCloneOptions() CloneOptions {
	return CloneOptions{
		MaxRetries:   3,
		ConnTimeout:  60 * time.Second,
		CloneTimeout: 10 * time.Minute,
		ProgressFunc: func(status string) {}, // No-op by default
		ExistingRepo: SkipExisting,
	}
}

// CloneResult represents the result of a clone operation
type CloneResult struct {
	RepoURL string
	Success bool
	Error   error
}

// ConcurrentCloner handles concurrent git clone operations
type ConcurrentCloner struct {
	maxConcurrent int
	semaphore     chan struct{}
	wg            sync.WaitGroup
}

// NewConcurrentCloner creates a new ConcurrentCloner
func NewConcurrentCloner(maxConcurrent int) *ConcurrentCloner {
	return &ConcurrentCloner{
		maxConcurrent: maxConcurrent,
		semaphore:     make(chan struct{}, maxConcurrent),
	}
}

// isGitRepo checks if a directory is a git repository
func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return false
	}
	return true
}

// handleExistingRepo handles an existing repository based on the strategy
func (c *ConcurrentCloner) handleExistingRepo(ctx context.Context, opts CloneOptions) error {
	if !isGitRepo(opts.TargetDir) {
		util.Debug(fmt.Sprintf("Target directory %s is not a git repository, proceeding with clone", opts.TargetDir))
		return nil // Not a git repo, proceed with clone
	}

	switch opts.ExistingRepo {
	case SkipExisting:
		util.Info(fmt.Sprintf("Skipping existing repository: %s", opts.URL))
		opts.ProgressFunc(fmt.Sprintf("Skipping existing repository: %s", opts.URL))
		return fmt.Errorf("repository already exists: %s", opts.TargetDir)

	case OverwriteExisting:
		util.Info(fmt.Sprintf("Removing existing repository: %s", opts.TargetDir))
		opts.ProgressFunc(fmt.Sprintf("Removing existing repository: %s", opts.TargetDir))
		if err := os.RemoveAll(opts.TargetDir); err != nil {
			util.Error("Failed to remove existing repository", err)
			return fmt.Errorf("failed to remove existing repository: %w", err)
		}
		return nil // Proceed with clone

	case FetchOnly:
		return c.fetchAndUpdate(ctx, opts)
	}

	return nil
}

// fetchAndUpdate updates an existing repository
func (c *ConcurrentCloner) fetchAndUpdate(ctx context.Context, opts CloneOptions) error {
	util.Info(fmt.Sprintf("Updating existing repository: %s", opts.URL))
	opts.ProgressFunc(fmt.Sprintf("Updating existing repository: %s", opts.URL))

	// Change to repository directory
	currentDir, err := os.Getwd()
	if err != nil {
		util.Error("Failed to get current directory", err)
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	if err := os.Chdir(opts.TargetDir); err != nil {
		util.Error("Failed to change to repository directory", err)
		return fmt.Errorf("failed to change to repository directory: %w", err)
	}
	defer os.Chdir(currentDir)

	// Fetch updates
	fetchCtx, cancel := context.WithTimeout(ctx, opts.ConnTimeout)
	defer cancel()
	fetchCmd := exec.CommandContext(fetchCtx, "git", "fetch", "--all", "--prune")
	if output, err := fetchCmd.CombinedOutput(); err != nil {
		util.Error("Failed to fetch updates", fmt.Errorf("%w: %s", err, output))
		return fmt.Errorf("failed to fetch updates: %w\nOutput: %s", err, output)
	}
	util.Debug("Successfully fetched updates")

	// Reset to specified branch or default branch
	resetCtx, cancel := context.WithTimeout(ctx, opts.ConnTimeout)
	defer cancel()
	resetArgs := []string{"reset", "--hard"}
	if opts.Branch != "" {
		resetArgs = append(resetArgs, fmt.Sprintf("origin/%s", opts.Branch))
	} else {
		resetArgs = append(resetArgs, "origin/HEAD")
	}
	resetCmd := exec.CommandContext(resetCtx, "git", resetArgs...)
	if output, err := resetCmd.CombinedOutput(); err != nil {
		util.Error("Failed to reset branch", fmt.Errorf("%w: %s", err, output))
		return fmt.Errorf("failed to reset branch: %w\nOutput: %s", err, output)
	}
	util.Debug("Successfully reset branch")

	util.Info(fmt.Sprintf("Successfully updated repository: %s", opts.URL))
	opts.ProgressFunc(fmt.Sprintf("Successfully updated repository: %s", opts.URL))
	return nil
}

// CloneRepository clones a single repository with retries and progress tracking
func (c *ConcurrentCloner) CloneRepository(ctx context.Context, opts CloneOptions) error {
	util.Info(fmt.Sprintf("Starting clone of repository: %s", opts.URL))

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(opts.TargetDir, 0755); err != nil {
		util.Error("Failed to create target directory", err)
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Handle existing repository
	if err := c.handleExistingRepo(ctx, opts); err != nil {
		if opts.ExistingRepo == SkipExisting {
			return nil // Skip is not an error condition
		}
		return err
	}

	var lastErr error
	for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate backoff duration (exponential)
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			msg := fmt.Sprintf("Retrying in %v... (attempt %d/%d)", backoff, attempt+1, opts.MaxRetries)
			util.Info(msg)
			opts.ProgressFunc(msg)
			time.Sleep(backoff)
		}

		// Set up command with timeouts
		cloneCtx, cancel := context.WithTimeout(ctx, opts.CloneTimeout)
		defer cancel()

		cmd := exec.CommandContext(cloneCtx, "git", "clone")
		if opts.Branch != "" {
			cmd.Args = append(cmd.Args, "-b", opts.Branch)
		}
		cmd.Args = append(cmd.Args, "--progress", opts.URL, opts.TargetDir)

		util.Debug(fmt.Sprintf("Running git command: %v", cmd.Args))

		// Capture command output
		output, err := cmd.CombinedOutput()
		if err == nil {
			msg := fmt.Sprintf("Successfully cloned %s", opts.URL)
			util.Info(msg)
			opts.ProgressFunc(msg)
			return nil
		}

		lastErr = fmt.Errorf("clone failed: %w\nOutput: %s", err, string(output))
		msg := fmt.Sprintf("Clone attempt %d failed: %v", attempt+1, lastErr)
		util.Error(msg, lastErr)
		opts.ProgressFunc(msg)
	}

	return fmt.Errorf("failed to clone after %d attempts: %w", opts.MaxRetries, lastErr)
}

// CloneRepositories clones multiple repositories concurrently
func (c *ConcurrentCloner) CloneRepositories(ctx context.Context, repos []CloneOptions) <-chan CloneResult {
	results := make(chan CloneResult, len(repos))

	go func() {
		defer close(results)

		util.Info(fmt.Sprintf("Starting concurrent clone of %d repositories", len(repos)))

		for _, opts := range repos {
			c.wg.Add(1)
			go func(opts CloneOptions) {
				defer c.wg.Done()

				// Acquire semaphore
				c.semaphore <- struct{}{}
				defer func() { <-c.semaphore }()

				err := c.CloneRepository(ctx, opts)
				result := CloneResult{
					RepoURL: opts.URL,
					Success: err == nil,
					Error:   err,
				}

				if result.Success {
					util.Info(fmt.Sprintf("Successfully cloned repository: %s", opts.URL))
				} else {
					util.Error(fmt.Sprintf("Failed to clone repository: %s", opts.URL), err)
				}

				results <- result
			}(opts)
		}

		c.wg.Wait()
		util.Info("Completed cloning all repositories")
	}()

	return results
}

// isDirEmpty checks if a directory is empty
func isDirEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == nil {
		return false, nil
	}
	if err != nil && err.Error() == "EOF" {
		return true, nil
	}
	return false, err
}
