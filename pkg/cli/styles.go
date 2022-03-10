package cli

import "github.com/charmbracelet/lipgloss"

var (
	Color   = lipgloss.AdaptiveColor{Light: "#111222", Dark: "#FAFAFA"}
	Primary = lipgloss.Color("#4636f5")
	Green   = lipgloss.Color("#9dcc3a")
	Red     = lipgloss.Color("#ff0000")
	White   = lipgloss.Color("#ffffff")
	Orange  = lipgloss.Color("#D3A347")

	Feint     = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#888888"}
	TextStyle = lipgloss.NewStyle().Foreground(Color)
	BoldStyle = TextStyle.Copy().Bold(true)
)

// RenderError returns a formatted error string.
func RenderError(msg string) string {
	// Error applies styles to an error message
	err := lipgloss.NewStyle().Background(Red).Foreground(White).Bold(true).Padding(0, 1).Render("Error")
	content := lipgloss.NewStyle().Bold(true).Padding(0, 1).Render(msg)
	return err + content
}

// RenderWarning returns a formatted warning string.
func RenderWarning(msg string) string {
	// Error applies styles to an error message
	err := lipgloss.NewStyle().Foreground(Orange).Bold(true).Render("Warning: ")
	content := lipgloss.NewStyle().Bold(true).Foreground(Orange).Padding(0, 1).Render(msg)
	return err + content
}
