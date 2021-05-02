package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	tree "github.com/marius/bubbles-tree"
)

const RootPath = tree.Path("/tmp")

func main() {
	path := RootPath
	if len(os.Args) > 1 {
		abs, err := filepath.Abs(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
		path = tree.Path(abs)
	}
	m := tree.New(path)
	m.Debug = true
	err := tea.NewProgram(m).Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s", err.Error())
	}
}
