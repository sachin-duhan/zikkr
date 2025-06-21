package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gh "github.com/sachin-duhan/zikrr/internal/github"
)

// View represents different screens in the TUI
type View int

const (
	ViewOrganization View = iota
	ViewRepositories
	ViewProgress
)

// Model represents the main TUI state
type Model struct {
	ctx         context.Context
	client      *gh.Client
	currentView View
	error       error

	// Window size
	width  int
	height int

	// View models
	organization *OrganizationModel
	repositories *RepositoriesModel
	progress     *ProgressModel

	// Shared state
	filter *gh.RepositoryFilter
}

// NewModel creates a new TUI model
func NewModel(ctx context.Context, client *gh.Client) Model {
	return Model{
		ctx:          ctx,
		client:       client,
		currentView:  ViewOrganization,
		filter:       &gh.RepositoryFilter{},
		organization: NewOrganizationModel(),
		repositories: NewRepositoriesModel(),
		progress:     NewProgressModel(),
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Handle view-specific updates
	switch m.currentView {
	case ViewOrganization:
		return m.updateOrganizationView(msg)
	case ViewRepositories:
		return m.updateRepositoriesView(msg)
	case ViewProgress:
		return m.updateProgressView(msg)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m Model) View() string {
	switch m.currentView {
	case ViewOrganization:
		return m.organizationView()
	case ViewRepositories:
		return m.repositoriesView()
	case ViewProgress:
		return m.progressView()
	default:
		return "Unknown view"
	}
}

// Common styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00")).
			MarginBottom(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00"))
)

// Error handling helper
func (m *Model) setError(err error) {
	if err != nil {
		m.error = fmt.Errorf("error: %w", err)
	} else {
		m.error = nil
	}
}

// renderError returns the error message if there is one
func (m Model) renderError() string {
	if m.error != nil {
		return errorStyle.Render(m.error.Error())
	}
	return ""
}

// SetOrganization pre-fills the organization name
func (m *Model) SetOrganization(org string) {
	if m.organization != nil {
		m.organization.input = org
	}
}
