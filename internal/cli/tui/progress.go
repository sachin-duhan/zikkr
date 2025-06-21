package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gh "github.com/sachin-duhan/zikrr/internal/github"
)

// progressBarWidth is the width of the progress bar
const progressBarWidth = 50

// ProgressModel represents the cloning progress view
type ProgressModel struct {
	cloneProgress map[string]float64
	totalProgress float64
	currentOp     string
	rateLimitInfo *gh.RateLimitInfo
	error         error
}

// NewProgressModel creates a new progress model
func NewProgressModel() *ProgressModel {
	return &ProgressModel{
		cloneProgress: make(map[string]float64),
	}
}

// updateProgressView handles updates for the progress view
func (m Model) updateProgressView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case progressMsg:
		m.progress.cloneProgress[msg.repo] = msg.progress
		m.progress.currentOp = msg.operation
		return m, nil

	case rateLimitMsg:
		m.progress.rateLimitInfo = msg.info
		return m, nil

	case errMsg:
		m.progress.error = msg.error
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			if m.progress.totalProgress == 100 {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

// progressView renders the cloning progress screen
func (m Model) progressView() string {
	var b strings.Builder

	// Title
	title := "Cloning Repositories"
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Overall progress
	total := 0.0
	for _, progress := range m.progress.cloneProgress {
		total += progress
	}
	m.progress.totalProgress = total / float64(len(m.repositories.selectedRepos)) * 100

	b.WriteString("Overall Progress:\n")
	b.WriteString(renderProgressBar(m.progress.totalProgress))
	b.WriteString(fmt.Sprintf(" %.1f%%\n\n", m.progress.totalProgress))

	// Current operation
	if m.progress.currentOp != "" {
		b.WriteString(fmt.Sprintf("Current Operation: %s\n\n", m.progress.currentOp))
	}

	// Individual repository progress
	b.WriteString("Repository Progress:\n")
	for repo, progress := range m.progress.cloneProgress {
		b.WriteString(fmt.Sprintf("%s:\n", repo))
		b.WriteString(renderProgressBar(progress))
		b.WriteString(fmt.Sprintf(" %.1f%%\n", progress))
	}

	// Rate limit info
	if m.progress.rateLimitInfo != nil {
		b.WriteString("\nGitHub API Rate Limit:\n")
		b.WriteString(fmt.Sprintf("Remaining: %d/%d\n", m.progress.rateLimitInfo.Remaining, m.progress.rateLimitInfo.Limit))
		if m.progress.rateLimitInfo.Remaining == 0 {
			resetTime := time.Until(m.progress.rateLimitInfo.Reset).Round(time.Second)
			b.WriteString(fmt.Sprintf("Reset in: %s\n", resetTime))
		}
	}

	// Instructions
	if m.progress.totalProgress == 100 {
		b.WriteString("\nAll repositories cloned successfully!\n")
		b.WriteString(infoStyle.Render("Press 'q' to quit"))
	}

	// Error message
	if m.progress.error != nil {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.progress.error.Error()))
	}

	return b.String()
}

// renderProgressBar creates a progress bar string
func renderProgressBar(percent float64) string {
	filled := int(percent / 100 * float64(progressBarWidth))
	if filled > progressBarWidth {
		filled = progressBarWidth
	}

	empty := progressBarWidth - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Render(fmt.Sprintf("[%s]", bar))
}

// Custom messages for progress updates
type (
	progressMsg struct {
		repo      string
		progress  float64
		operation string
	}

	rateLimitMsg struct {
		info *gh.RateLimitInfo
	}
)
