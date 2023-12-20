package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tree "github.com/mariusor/bubbles-tree"
)

type message struct {
	viewport.Model
	count    int
	level    int
	state    tree.NodeState
	parent   tree.Node
	children []*message
}

func (m *message) setChildren(nodes ...tree.Node) {
	m.children = m.children[:0]
	for _, nn := range nodes {
		if c, ok := nn.(*message); ok {
			m.children = append(m.children, c)
		}
	}
}

func (m *message) Init() tea.Cmd {
	m.state = tree.NodeIsMultiLine
	return nil
}

func (m *message) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	switch mm := msg.(type) {
	case tree.NodeState:
		m.state |= mm
	case tree.Nodes:
		m.setChildren(mm...)
	}
	return m, cmd
}

func (m *message) View() string {
	return m.Model.View()
}

func (m *message) Parent() tree.Node {
	return m.parent
}

func (m *message) Children() tree.Nodes {
	return treeNodes(m.children)
}

func level(p tree.Node) int {
	if p == nil {
		return 0
	}
	lvl := 0
	for {
		if p = p.Parent(); p == nil {
			break
		}
		lvl++
	}
	return lvl
}

func treeNodes(pathNodes []*message) tree.Nodes {
	nodes := make(tree.Nodes, len(pathNodes))
	for i, n := range pathNodes {
		nodes[i] = n
	}
	return nodes
}

func (m *message) State() tree.NodeState {
	state := m.state
	if len(m.children) > 0 || m.Model.Height > 0 {
		state |= tree.NodeCollapsible
	}
	return state
}

var _ tree.Node = new(message)

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

func buildMessage(parent tree.Node, depth, count int) message {
	t := viewport.New(0, 0)

	m := message{Model: t, parent: parent, count: count, level: level(parent) + 1}
	m.children = buildConversation(depth-1, &m)

	bold := lipgloss.NewStyle().Bold(true)
	var title string
	if parent == nil {
		title = bold.Render("Root node")
	} else {
		title = bold.Render(fmt.Sprintf("Child node #%d-%d", m.level, m.count))
	}
	lipsum := lipgloss.JoinVertical(lipgloss.Top, title, "Sphinx of black quartz, judge my vow!\nThe quick brown fox jumps over the lazy dog.")
	m.Model.Height = lipgloss.Height(lipsum) + 2
	m.Model.SetContent(lipsum)
	m.Model.Style = m.Model.Style.Foreground(lipgloss.Color("silver")).PaddingTop(1).PaddingBottom(1)

	return m
}

func buildConversation(depth int, parent tree.Node) []*message {
	if depth == 0 {
		return nil
	}
	conv := make([]*message, 0)
	maxMessages := 0
	for {
		if maxMessages = rand.Intn(10); maxMessages > 0 {
			break
		}
	}

	for i := 0; i < maxMessages; i++ {
		m := buildMessage(parent, depth, i)
		conv = append(conv, &m)
	}
	return conv
}

type depthStyle struct {
	lipgloss.Style
	colors []lipgloss.TerminalColor
}

func (ds depthStyle) Width(w int) tree.DepthStyler {
	ds.Style = ds.Style.Width(w)
	return ds
}

func (ds depthStyle) Render(d int, strs ...string) string {
	d = d % len(ds.colors)
	ds.Style = ds.Style.Foreground(ds.colors[d])
	return ds.Style.Render(strs...)
}

func main() {
	var depth int
	flag.IntVar(&depth, "depth", 2, "The depth to which to build the conversation")
	flag.Parse()

	t := tree.New(treeNodes(buildConversation(depth, nil)))
	t.Symbols = tree.ThickEdgeSymbols()
	t.Styles.Selected = t.Styles.Line

	t.Styles.Symbol = depthStyle{
		Style: lipgloss.NewStyle(),
		colors: []lipgloss.TerminalColor{
			lipgloss.Color("#ff0000"),
			lipgloss.Color("#00ff00"),
			lipgloss.Color("#0000ff"),
			lipgloss.Color("#00ffff"),
			lipgloss.Color("#ff00ff"),
			lipgloss.Color("#ffff00"),
		},
	}
	m := quittingTree{Model: t}

	if _, err := tea.NewProgram(&m).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
