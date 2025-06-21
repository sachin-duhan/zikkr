package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sachin-duhan/zikrr/internal/git"
)

var (
	statusColors = map[git.RepositoryStatus]lipgloss.Style{
		git.StatusPending:  lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		git.StatusCloning:  lipgloss.NewStyle().Foreground(lipgloss.Color("33")),
		git.StatusRetrying: lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		git.StatusSuccess:  lipgloss.NewStyle().Foreground(lipgloss.Color("42")),
		git.StatusFailed:   lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		git.StatusSkipped:  lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		git.StatusUpdating: lipgloss.NewStyle().Foreground(lipgloss.Color("99")),
	}

	strategyNames = map[git.ExistingRepoStrategy]string{
		git.SkipExisting:      "Skip",
		git.OverwriteExisting: "Overwrite",
		git.FetchOnly:         "Update",
	}
)

// ProgressModel represents the progress view state
type ProgressModel struct {
	repoManager *git.RepositoryManager
	progress    progress.Model
	width       int
	height      int
	done        bool
	err         error
	updates     <-chan *git.Repository
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewProgressModel creates a new progress model
func NewProgressModel(baseDir string, maxConcurrent int) *ProgressModel {
	ctx, cancel := context.WithCancel(context.Background())
	return &ProgressModel{
		repoManager: git.NewRepositoryManager(baseDir, maxConcurrent),
		progress:    progress.New(progress.WithDefaultGradient()),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// AddRepository adds a repository to be cloned
func (m *ProgressModel) AddRepository(org, name, url, branch string, strategy git.ExistingRepoStrategy) {
	m.repoManager.AddRepository(org, name, url, branch, strategy)
}

// StartCloning starts the cloning process
func (m *ProgressModel) StartCloning() tea.Cmd {
	return func() tea.Msg {
		m.updates = m.repoManager.CloneAll(m.ctx)
		return nil
	}
}

// Update handles model updates
func (m *ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 4
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancel()
			return m, tea.Quit
		}
	}

	if m.updates != nil {
		select {
		case repo, ok := <-m.updates:
			if !ok {
				m.done = true
				return m, tea.Quit
			}
			if repo != nil {
				return m, nil
			}
		default:
		}
	}

	prog, cmd := m.progress.Update(msg)
	m.progress = prog.(progress.Model)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the progress view
func (m *ProgressModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	var s strings.Builder
	s.WriteString("\n  Cloning Repositories\n\n")

	// Show repository status
	repos := m.repoManager.GetRepositories()
	total := len(repos)
	completed := 0
	skipped := 0
	failed := 0

	for _, repo := range repos {
		status, err, progress := repo.GetStatus()
		statusStyle := statusColors[status]

		// Format repository line
		repoLine := fmt.Sprintf("  %s/%s", repo.Organization, repo.Name)

		// Add strategy for existing repos if relevant
		if status == git.StatusSkipped || status == git.StatusUpdating {
			repoLine += fmt.Sprintf(" [%s]", strategyNames[repo.ExistingRepo])
		}

		// Add progress or error information
		if status == git.StatusCloning && progress != "" {
			repoLine += fmt.Sprintf(" - %s", progress)
		} else if status == git.StatusUpdating && progress != "" {
			repoLine += fmt.Sprintf(" - %s", progress)
		}
		if err != nil {
			repoLine += fmt.Sprintf(" - Error: %v", err)
		}

		s.WriteString(statusStyle.Render(repoLine) + "\n")

		// Update counters
		switch status {
		case git.StatusSuccess:
			completed++
		case git.StatusSkipped:
			skipped++
		case git.StatusFailed:
			failed++
		}
	}

	// Show overall progress
	s.WriteString("\n")
	if total > 0 {
		progress := float64(completed+skipped) / float64(total)
		s.WriteString(fmt.Sprintf("  %s\n", m.progress.ViewAs(progress)))
		s.WriteString(fmt.Sprintf("  Progress: %d/%d repositories\n", completed+skipped, total))
		s.WriteString(fmt.Sprintf("  • Completed: %d\n", completed))
		s.WriteString(fmt.Sprintf("  • Skipped: %d\n", skipped))
		if failed > 0 {
			s.WriteString(fmt.Sprintf("  • Failed: %d\n", failed))
		}
	}

	// Show completion message
	if m.done {
		s.WriteString("\n  Done! Press q to exit\n")
	}

	return s.String()
}

// Init initializes the model
func (m *ProgressModel) Init() tea.Cmd {
	return tea.Batch(
		m.progress.Init(),
		m.StartCloning(),
	)
}
