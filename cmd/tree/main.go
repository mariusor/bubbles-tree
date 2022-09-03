package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	tree "github.com/mariusor/bubbles-tree"
)

const RootPath = tree.Path("/tmp")

type example struct {
	tree.Model
}

func (e *example) Update(m tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := m.(tea.KeyMsg); ok && key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		e.Model.LogFn("exiting")
		return e, tea.Quit
	}
	_, cmd := e.Model.Update(m)
	return e, cmd
}

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
	m := example{tree.New(path)}
	m.Debug = true
	out, err := os.OpenFile(filepath.Join("/tmp", os.Args[0]+".log"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err == nil {
		log.SetOutput(out)
		m.LogFn = log.Printf
	}

	if err := tea.NewProgram(&m).Start(); err != nil {
		log.Printf("Err: %s", err.Error())
	}
}
