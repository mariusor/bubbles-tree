package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	tree "github.com/marius/bubbles-tree"
)

const RootPath = tree.Path("/tmp")

func main() {
	path := RootPath
	if len(os.Args) > 1 {
		path = tree.Path(os.Args[1])
	}
	m := tree.New(path)
	m.Debug = true
	err := tea.NewProgram(m).Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s", err.Error())
	}
}
