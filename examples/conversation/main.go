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
	s := strings.Builder{}
	fmt.Fprintf(&s, "\n%s\n", m.Model.View())
	return s.String()
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

type coloredBorder []lipgloss.TerminalColor

var pipe = "  │"

func (c coloredBorder) Padding(_ int) string {
	return "  "
}

func (c coloredBorder) DrawNode(d int) string {
	return c.drawPipe(d)
}

func (c coloredBorder) DrawLast(d int) string {
	return c.drawPipe(d)
}

func (c coloredBorder) DrawVertical(d int) string {
	return c.drawPipe(d)
}

func (c coloredBorder) drawPipe(d int) string {
	d = d % len(c)
	style := lipgloss.Style{}
	style = style.Foreground(c[d])
	return style.Render(pipe)
}

var _ tree.DrawSymbols = &coloredBorder{}

func main() {
	var depth int
	flag.IntVar(&depth, "depth", 2, "The depth to which to build the conversation")
	flag.Parse()

	t := tree.New(buildConversation(depth, nil))
	t.Symbols = coloredBorder{
		lipgloss.Color("#ff0000"),
		lipgloss.Color("#00ff00"),
		lipgloss.Color("#0000ff"),
		lipgloss.Color("#00ffff"),
		lipgloss.Color("#ff00ff"),
		lipgloss.Color("#ffff00"),
	}
	m := quittingTree{Model: t}

	if _, err := tea.NewProgram(&m).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}