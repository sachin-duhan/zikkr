package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v60/github"
)

const reposPerPage = 10

// RepositoriesModel represents the repository selection view
type RepositoriesModel struct {
	repositories  []*github.Repository
	selectedRepos map[string]bool
	cursor        int
	page          int
	totalPages    int
	filterVisible bool
	error         error
}

// NewRepositoriesModel creates a new repositories model
func NewRepositoriesModel() *RepositoriesModel {
	return &RepositoriesModel{
		selectedRepos: make(map[string]bool),
	}
}

// SetRepositories updates the repositories list and recalculates pages
func (r *RepositoriesModel) SetRepositories(repos []*github.Repository) {
	r.repositories = repos
	r.totalPages = (len(repos) + reposPerPage - 1) / reposPerPage
	r.page = 0
	r.cursor = 0
}

// GetPageRepos returns the repositories for the current page
func (r *RepositoriesModel) GetPageRepos() []*github.Repository {
	start := r.page * reposPerPage
	end := start + reposPerPage
	if end > len(r.repositories) {
		end = len(r.repositories)
	}
	if start >= len(r.repositories) {
		return nil
	}
	return r.repositories[start:end]
}

// updateRepositoriesView handles updates for the repository selection view
func (m Model) updateRepositoriesView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case reposMsg:
		m.repositories.SetRepositories(msg.repos)
		return m, nil

	case errMsg:
		m.repositories.error = msg.error
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.repositories.cursor > 0 {
				m.repositories.cursor--
			}
		case "down", "j":
			maxCursor := len(m.repositories.GetPageRepos()) - 1
			if m.repositories.cursor < maxCursor {
				m.repositories.cursor++
			}
		case "left", "h":
			if m.repositories.page > 0 {
				m.repositories.page--
				m.repositories.cursor = 0
			}
		case "right", "l":
			if m.repositories.page < m.repositories.totalPages-1 {
				m.repositories.page++
				m.repositories.cursor = 0
			}
		case " ":
			// Toggle selection
			repos := m.repositories.GetPageRepos()
			if len(repos) > m.repositories.cursor {
				repo := repos[m.repositories.cursor]
				fullName := repo.GetFullName()
				m.repositories.selectedRepos[fullName] = !m.repositories.selectedRepos[fullName]
			}
		case "f":
			m.repositories.filterVisible = !m.repositories.filterVisible
		case "enter":
			if len(m.repositories.selectedRepos) > 0 {
				m.currentView = ViewProgress
				return m, m.startCloning
			}
		}
	}
	return m, nil
}

// repositoriesView renders the repository selection screen
func (m Model) repositoriesView() string {
	var b strings.Builder

	// Title
	title := fmt.Sprintf("%s - Select Repositories", m.organization.name)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Repository list
	repos := m.repositories.GetPageRepos()
	for i, repo := range repos {
		cursor := " "
		if m.repositories.cursor == i {
			cursor = ">"
		}

		selected := " "
		if m.repositories.selectedRepos[repo.GetFullName()] {
			selected = "✓"
		}

		// Repository info
		repoInfo := fmt.Sprintf(
			"%s [%s] %s (%d ⭐️, %s)",
			cursor,
			selected,
			repo.GetFullName(),
			repo.GetStargazersCount(),
			repo.GetLanguage(),
		)

		// Style based on cursor position
		if m.repositories.cursor == i {
			repoInfo = cursorStyle.Render(repoInfo)
		}
		if m.repositories.selectedRepos[repo.GetFullName()] {
			repoInfo = selectedStyle.Render(repoInfo)
		}

		b.WriteString(repoInfo)
		b.WriteString("\n")
	}

	// Pagination info
	pageInfo := fmt.Sprintf("\nPage %d/%d", m.repositories.page+1, m.repositories.totalPages)
	b.WriteString(infoStyle.Render(pageInfo))
	b.WriteString("\n\n")

	// Instructions
	instructions := []string{
		"↑/k, ↓/j: Navigate",
		"←/h, →/l: Change page",
		"Space: Toggle selection",
		"f: Toggle filters",
		"Enter: Start cloning",
		"q: Quit",
	}
	for _, instruction := range instructions {
		b.WriteString(infoStyle.Render(instruction))
		b.WriteString("\n")
	}

	// Selection summary
	summary := fmt.Sprintf("\nSelected: %d repositories", len(m.repositories.selectedRepos))
	b.WriteString(infoStyle.Render(summary))

	// Error message
	if m.repositories.error != nil {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.repositories.error.Error()))
	}

	return b.String()
}

// startCloning is a command that initiates the cloning process
func (m Model) startCloning() tea.Msg {
	// This will be implemented in the next phase
	return nil
}
