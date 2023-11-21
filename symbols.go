package tree

import "strings"

type Symbols struct {
	// We should try to copy the API of lipgloss.Border
	Width int

	Continuator string
	Starter     string
	Terminator  string
	Horizontal  string

	Collapsed string
	Expanded  string
}

// Padding is expected to output a whitespace, or equivalent, used when two nodes
// at the same level are not children to the same parent.
func Padding(style Style, s Symbols, _ int) string {
	return strings.Repeat(" ", s.Width)
}

// RenderTerminator is expected to output a terminator marker used for the last node in a list of nodes.
func RenderTerminator(style Style, s Symbols, _ int) string {
	return draw(style, s.Terminator, s.Width)
}

// RenderStarter is expected to output the marker used for every node in the tree.
func RenderStarter(style Style, s Symbols, _ int) string {
	return draw(style, s.Starter, s.Width)
}

// RenderContinuator is expected to output a continuator marker used to connect two nodes
// which are children on the same parent.
func RenderContinuator(style Style, s Symbols, _ int) string {
	return draw(style, s.Continuator, s.Width)
}

// DefaultSymbols returns a set of default Symbols for drawing the tree.
func DefaultSymbols() Symbols {
	return normalSymbols
}

var (
	normalSymbols = Symbols{
		Width:       3,
		Starter:     "├─",
		Continuator: "│ ",
		Terminator:  "└─",
	}

	roundedSymbols = Symbols{
		Width:       3,
		Starter:     "├─",
		Continuator: "│ ",
		Terminator:  "╰─",
	}

	thickSymbols = Symbols{
		Width:       3,
		Starter:     "┣━",
		Continuator: "┃ ",
		Terminator:  "┗━",
	}

	doubleSymbols = Symbols{
		Width:       3,
		Starter:     "╠═",
		Continuator: "║",
		Terminator:  "╚═",
	}

	thickEdgeSymbols = Symbols{
		Width:       2,
		Starter:     " ╷",
		Continuator: " │",
		Terminator:  " ╵",
	}

	normalEdgeSymbols = Symbols{
		Width:       2,
		Starter:     " ╻",
		Continuator: " ┃",
		Terminator:  " ╹",
	}
)

// NormalSymbols returns a standard-type symbols with a normal weight and 90
// degree corners.
func NormalSymbols() Symbols {
	return normalSymbols
}

// RoundedSymbols returns a symbols with rounded corners.
func RoundedSymbols() Symbols {
	return roundedSymbols
}

// ThickSymbols returns a symbols that's thicker than the one returned by
// NormalSymbols.
func ThickSymbols() Symbols {
	return thickSymbols
}

// DoubleSymbols returns a symbols comprised of two thin strokes.
func DoubleSymbols() Symbols {
	return doubleSymbols
}

func ThickEdgeSymbols() Symbols {
	return thickEdgeSymbols
}
