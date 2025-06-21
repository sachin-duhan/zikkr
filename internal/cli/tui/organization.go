package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v60/github"
)

// OrganizationModel represents the organization input view
type OrganizationModel struct {
	input string
	name  string
	error error
}

// NewOrganizationModel creates a new organization model
func NewOrganizationModel() *OrganizationModel {
	return &OrganizationModel{}
}

// updateOrganizationView handles updates for the organization input view
func (m Model) updateOrganizationView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if strings.TrimSpace(m.organization.input) != "" {
				m.organization.name = strings.TrimSpace(m.organization.input)
				m.currentView = ViewRepositories
				return m, m.fetchRepositories
			}
		case tea.KeyBackspace:
			if len(m.organization.input) > 0 {
				m.organization.input = m.organization.input[:len(m.organization.input)-1]
			}
		case tea.KeyRunes:
			m.organization.input += string(msg.Runes)
		}
	}
	return m, nil
}

// organizationView renders the organization input screen
func (m Model) organizationView() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render("Zikrr - GitHub Organization Cloner")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Input prompt
	prompt := "Enter GitHub organization name: "
	b.WriteString(prompt)

	// Input field
	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#333333")).
		Padding(0, 1)

	input := m.organization.input
	if input == "" {
		input = " " // Show empty input field
	}
	b.WriteString(inputStyle.Render(input))
	b.WriteString("\n\n")

	// Instructions
	instructions := []string{
		"Press Enter to continue",
		"Press Ctrl+C to quit",
	}
	for _, instruction := range instructions {
		b.WriteString(infoStyle.Render(instruction))
		b.WriteString("\n")
	}

	// Error message
	if m.organization.error != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(m.organization.error.Error()))
	}

	return b.String()
}

// fetchRepositories is a command that fetches repositories for the organization
func (m Model) fetchRepositories() tea.Msg {
	repos, err := m.client.ListFilteredRepositories(m.ctx, m.organization.name, m.filter)
	if err != nil {
		return errMsg{err}
	}
	return reposMsg{repos}
}

// Custom messages
type (
	errMsg struct {
		error
	}

	reposMsg struct {
		repos []*github.Repository
	}
)
