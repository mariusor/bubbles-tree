package main

import (
	"flag"
	"fmt"
	"io"
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
	var debug bool
	var depth int
	flag.IntVar(&depth, "depth", 2, "The maximum depth to read the directory structure")
	flag.BoolVar(&debug, "debug", false, "Are we debugging")
	flag.Parse()

	log.SetOutput(openlog())

	path := RootPath
	if flag.NArg() > 0 {
		abs, err := filepath.Abs(flag.Arg(0))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}
		path = abs
	}

	log.Printf("starting at %s", path)
	t := tree.New(buildNodeTree(path, 2))
	m := quittingTree{Model: t}

	m.Model.LogFn = log.Printf

	initializers := make([]tea.ProgramOption, 0)
	if debug {
		nilReader := io.LimitedReader{}
		initializers = append(initializers /*tea.WithoutRenderer(), */, tea.WithInput(&nilReader))
	}
	if err := tea.NewProgram(&m, initializers...).Start(); err != nil {
		log.Printf("Err: %s", err.Error())
	}
}