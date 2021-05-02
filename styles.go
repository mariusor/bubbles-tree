package tree

import "github.com/charmbracelet/lipgloss"

const (
	BoxDrawingsVerticalAndRight = "├"
	BoxDrawingsVertical = "│"
	BoxDrawingsUpAndRight = "└"
	BoxDrawingsDownAndRight = "┌"
	BoxDrawingsHorizontal = "─"

	SquaredPlus = "⊞"
	SquaredMinus = "⊟"
)

var (
	DefaultStyle   = lipgloss.Style{}

	errForegroundColor = lipgloss.AdaptiveColor{Light: "#E03F3F", Dark: "#F45B5B"}
	errBackgroundColor = lipgloss.AdaptiveColor{Light: "#212121", Dark: "#4A4A4A"}
	ErrStyle           = lipgloss.Style{}.Foreground(errForegroundColor).Background(errBackgroundColor)

	debugForegroundColor = lipgloss.AdaptiveColor{Light: "#FFA348", Dark: "#FFBE6F"}
	debugBackgroundColor = lipgloss.AdaptiveColor{Light: "#212121", Dark: "#4A4A4A"}
	DebugStyle           = lipgloss.Style{}.Foreground(debugForegroundColor).Background(debugBackgroundColor)
)
