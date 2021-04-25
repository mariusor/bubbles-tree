package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	tree "github.com/marius/bubbles-tree"
)

func main() {
	m := new(tree.Model)
	m.Root = "/tmp"
	err := tea.NewProgram(m).Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s", err.Error())
	}
}
