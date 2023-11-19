package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tree "github.com/mariusor/bubbles-tree"
	"gopkg.in/loremipsum.v1"
)

type message struct {
	textarea.Model
	parent   tree.Node
	children tree.Nodes
}

func (m message) Init() tea.Cmd {
	return nil
}

func (m message) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m message) View() string {
	return m.Model.View()
}

func (m message) Parent() tree.Node {
	return m.parent
}

func (m message) Children() tree.Nodes {
	return m.children
}

func (m message) State() tree.NodeState {
	var state tree.NodeState = 0
	if len(m.children) > 0 {
		state |= tree.NodeCollapsible
	}
	return state
}

var _ tree.Node = message{}

type quittingTree struct {
	tree.Model
}

func (e *quittingTree) Update(m tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := m.(tea.KeyMsg); ok && key.Matches(msg, key.NewBinding(key.WithKeys("q"))) {
		return e, tea.Quit
	}
	_, cmd := e.Model.Update(m)
	return e, cmd
}

func buildMessage(parent tree.Node, depth int) message {
	m := message{
		Model:  textarea.New(),
		parent: parent,
	}
	m.Model.ShowLineNumbers = false
	m.Model.SetValue(loremipsum.New().Paragraph())
	m.Model.SetPromptFunc(1, func(lineIdx int) string { return "" })
	m.children = buildConversation(depth-1, &m)
	return m
}

func buildConversation(depth int, parent tree.Node) tree.Nodes {
	if depth == 0 {
		return nil
	}
	conv := make(tree.Nodes, 0)
	maxMessages := 0
	for {
		if maxMessages = rand.Intn(3); maxMessages > 0 {
			break
		}
	}

	for i := 0; i < maxMessages; i++ {
		conv = append(conv, buildMessage(parent, depth))
	}
	return conv
}

type coloredBorder struct {
	depth int
}

var pipe = " " + lipgloss.NormalBorder().Left

func (c coloredBorder) Padding() string {
	return strings.Repeat(" ", 3)
}

func (c coloredBorder) DrawNode() string {
	return pipe
}

func (c coloredBorder) DrawLast() string {
	return pipe + "\n"
}

func (c coloredBorder) DrawVertical() string {
	return pipe
}

var _ tree.DrawSymbols = coloredBorder{}

func main() {
	var depth int
	flag.IntVar(&depth, "depth", 2, "The depth to which to build the conversation")
	flag.Parse()

	t := tree.New(buildConversation(depth, nil))
	t.Symbols = coloredBorder{}
	m := quittingTree{Model: t}

	if _, err := tea.NewProgram(&m).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
