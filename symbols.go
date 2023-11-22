package tree

import "unicode/utf8"

type Symbols struct {
	Connector  string
	Starter    string
	Terminator string
	Horizontal string
}

func width(s Symbols) int {
	co := utf8.RuneCount([]byte(s.Connector))
	st := utf8.RuneCount([]byte(s.Starter))
	te := utf8.RuneCount([]byte(s.Terminator))
	return max(max(co, st), te) + 1
}

// Padding is expected to output a whitespace, or equivalent, used when two nodes
// at the same level are not children to the same parent.
func Padding(style DepthStyler, s Symbols, depth int) string {
	return draw(style, " ", width(s), depth)
}

// RenderTerminator is expected to output a terminator marker used for the last node in a list of nodes.
func RenderTerminator(style DepthStyler, s Symbols, depth int) string {
	return draw(style, s.Terminator, width(s), depth)
}

// RenderStarter is expected to output the marker used for every node in the tree.
func RenderStarter(style DepthStyler, s Symbols, depth int) string {
	return draw(style, s.Starter, width(s), depth)
}

// RenderConnector is expected to output a continuator marker used to connect two nodes
// which are children on the same parent.
func RenderConnector(style DepthStyler, s Symbols, depth int) string {
	return draw(style, s.Connector, width(s), depth)
}

// DefaultSymbols returns a set of default Symbols for drawing the tree.
func DefaultSymbols() Symbols {
	return normalSymbols
}

var (
	normalSymbols = Symbols{
		Starter:    "├─",
		Connector:  "│ ",
		Terminator: "└─",
	}

	roundedSymbols = Symbols{
		Starter:    "├─",
		Connector:  "│ ",
		Terminator: "╰─",
	}

	thickSymbols = Symbols{
		Starter:    "┣━",
		Connector:  "┃ ",
		Terminator: "┗━",
	}

	doubleSymbols = Symbols{
		Starter:    "╠═",
		Connector:  "║",
		Terminator: "╚═",
	}

	normalEdgeSymbols = Symbols{
		Starter:    "╷",
		Connector:  "│",
		Terminator: "╵",
	}

	thickEdgeSymbols = Symbols{
		Starter:    "╻",
		Connector:  "┃",
		Terminator: "╹",
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

func NormalEdgeSymbols() Symbols {
	return normalEdgeSymbols
}

func ThickEdgeSymbols() Symbols {
	return thickEdgeSymbols
}
