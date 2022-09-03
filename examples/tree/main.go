package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	tree "github.com/mariusor/bubbles-tree"
)

const RootPath = "/tmp"

type quittingTree struct {
	tree.Model
}

func New(fs fs.FS) quittingTree {
	t := tree.New(&dirFs{fs})
	return quittingTree{Model: t}
}

func (e *quittingTree) Update(m tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := m.(tea.KeyMsg); ok && key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		e.Model.LogFn("exiting")
		return e, tea.Quit
	}
	_, cmd := e.Model.Update(m)
	return e, cmd
}

func openlog() io.Writer {
	name := filepath.Join("/tmp", filepath.Base(os.Args[0])+".log")
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return io.Discard
	}
	return f
}

func main() {
	path := RootPath
	if len(os.Args) > 1 {
		abs, err := filepath.Abs(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
		path = abs
	}
	m := New(os.DirFS(path))

	log.SetOutput(openlog())
	m.Model.LogFn = log.Printf

	if err := tea.NewProgram(&m).Start(); err != nil {
		log.Printf("Err: %s", err.Error())
	}
}
