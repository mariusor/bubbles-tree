package tree

import "strings"

// DrawSymbols represents an uniform API to be used when drawing the tree elements.
type DrawSymbols interface {
	// Padding is expected to output a whitespace, or equivalent, used when two nodes
	// at the same level are not children to the same parent.
	Padding(int) string
	// DrawNode is expected to output the marker used for every node in the tree.
	DrawNode(int) string
	// DrawLast is expected to output a terminator marker used for the last node in a list of nodes.
	DrawLast(int) string
	// DrawVertical is expected to output a continuator marker used to connect two nodes
	// which are children on the same parent.
	DrawVertical(int) string
}

type Symbols struct {
	// We should try to copy the API of lipgloss.Symbolss
	Width int

	Vertical         Symbol
	VerticalAndRight Symbol
	UpAndRight       Symbol
	Horizontal       Symbol

	Collapsed string
	Expanded  string
	Ellipsis  string
}

func (s Symbols) Padding(_ int) string {
	return strings.Repeat(" ", s.Width)
}

func (s Symbols) DrawLast(_ int) string {
	return s.UpAndRight.draw(s.Width)
}

func (s Symbols) DrawNode(_ int) string {
	return s.VerticalAndRight.draw(s.Width)
}

func (s Symbols) DrawVertical(_ int) string {
	return s.Vertical.draw(s.Width)
}

// DefaultSymbols returns a set of default Symbols for drawing the tree.
func DefaultSymbols() DrawSymbols {
	return normalSymbols
}

var (
	normalSymbols = Symbols{
		Width:            3,
		Vertical:         "│ ",
		VerticalAndRight: "├─",
		UpAndRight:       "└─",

		Ellipsis: "…",
	}

	roundedSymbols = Symbols{
		Width:            3,
		Vertical:         "│ ",
		VerticalAndRight: "├─",
		UpAndRight:       "╰─",

		Ellipsis: "…",
	}

	thickSymbols = Symbols{
		Width:            3,
		Vertical:         "┃ ",
		VerticalAndRight: "┣━",
		UpAndRight:       "┗━",

		Ellipsis: "…",
	}

	doubleSymbols = Symbols{
		Width:            3,
		Vertical:         "║",
		VerticalAndRight: "╠═",
		UpAndRight:       "╚═",

		Ellipsis: "…",
	}
)

// NormalSymbols returns a standard-type symbols with a normal weight and 90
// degree corners.
func NormalSymbols() DrawSymbols {
	return normalSymbols
}

// RoundedSymbols returns a symbols with rounded corners.
func RoundedSymbols() DrawSymbols {
	return roundedSymbols
}

// ThickSymbols returns a symbols that's thicker than the one returned by
// NormalSymbols.
func ThickSymbols() DrawSymbols {
	return thickSymbols
}

// DoubleSymbols returns a symbols comprised of two thin strokes.
func DoubleSymbols() DrawSymbols {
	return doubleSymbols
}
