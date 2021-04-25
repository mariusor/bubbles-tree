package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	tree "github.com/marius/bubbles-tree"
)

func main() {
	err := tea.NewProgram(tree.New(tree.Path("/tmp"))).Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s", err.Error())
	}
}
