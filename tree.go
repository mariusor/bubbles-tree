package tree

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type NodeState int

const (
	NodeCollapsed = 1 << iota
	NodeCollapsible
	NodeError
	NodeDebug
)

type Treeish interface {
	Advance(string) Treeish
	Walk(int) ([]string, error)
	State(string) (NodeState, error)
}

type debugNode struct {
	Content interface{}
	state   NodeState
}

func debug(m interface{}) debugNode {
	return debugNode{Content: m}
}

func (m *Model) debug(s string, params ...interface{}) {
	if m.debugNodes == nil {
		m.debugNodes = make([]Node, 0)
	}
	m.debugNodes = append(m.debugNodes, debug(fmt.Sprintf(s, params...)))
}

func (d debugNode) String() string {
	switch n := d.Content.(type) {
	case error:
		return n.Error()
	case string:
		return n
	}
	return fmt.Sprintf("unknown type %T", d)
}

func (d debugNode) Children() Nodes {
	return nil
}

func (d debugNode) State() NodeState {
	if _, ok := d.Content.(error); ok {
		return NodeError
	}
	return NodeDebug
}

type pathNode struct {
	Path  string
	state NodeState
	Nodes Nodes
}

func (n pathNode) GoString() string {
	return fmt.Sprintf("Node(%s)[%d]\n", n.Path, len(n.Children()))
}

func (n pathNode) String() string {
	return n.Path
}

func (n pathNode) Children() Nodes {
	return n.Nodes
}

func (n pathNode) State() NodeState {
	return n.state
}

type Node interface {
	String() string
	Children() Nodes
	State() NodeState
}

type viewport struct {
	h, w      int
	top, left int
	pos       int
}

type Nodes []Node

func (n Nodes) Len() int {
	len := 0
	for _, node := range n {
		len++
		if node.Children() != nil {
			len += node.Children().Len()
		}
	}
	return len
}

// Model is the Bubble Tea model for this user interface.
type Model struct {
	tree       Nodes
	debugNodes Nodes
	showDebug  bool

	t    Treeish
	view viewport
}

func New(t Treeish) *Model {
	m := new(Model)
	m.t = t
	return m
}

func (m *Model) Nodes() Nodes {
	bot := min(m.view.top+m.view.h, len(m.tree))
	//top := clamp(m.view.top, 0, len(m.tree)-bot)
	m.debugNodes = append(m.debugNodes, debug(fmt.Sprintf("h:%d p:%d :: t:%d b:%d l:%d\n", m.view.h, m.view.pos, m.view.top, bot, len(m.tree))))
	if m.showDebug {
		return m.debugNodes
	}
	return m.tree //[top:bot]
}

func (m Model) nodeAt(i int) Node {
	for j, p := range m.tree {
		if j == i {
			return p
		}
	}
	return nil
}

// ToggleExpand
func (m *Model) ToggleExpand() error {
	if pn, ok := m.nodeAt(m.view.pos).(*pathNode); ok {
		pn.state ^= NodeCollapsed
	}
	return nil
}

// Prev moves the current position to the previous 'i'th element in the tree.
// If it's above the viewport we need to recompute the top
func (m *Model) Prev(i int) error {
	m.view.pos = clamp(m.view.pos-i, 0, m.view.h)
	if m.view.pos < m.view.top {
		m.view.top = clamp(m.view.pos, 0, m.view.h)
	}
	return nil
}

// Next moves the current position to the next 'i'th element in the tree
// If it's below the viewport we need to recompute the top
func (m *Model) Next(i int) error {
	m.view.pos = clamp(m.view.pos+i, 0, min(m.view.h, m.tree.Len()))
	bot := min(m.view.top+m.view.h-1, m.view.h)
	if m.view.pos > bot {
		m.view.top = clamp(bot+i, 0, m.view.h)
	}
	return nil
}

type TreeMsg string

func (m *Model) init() tea.Msg {
	walk(m)
	return TreeMsg("inited")
}

func (m *Model) Init() tea.Cmd {
	return m.init
}

func findNodeByPath(nodes Nodes, path string) Node {
	for _, node := range nodes {
		if filepath.Clean(node.String()) == filepath.Clean(path) {
			return node
		}
		child := findNodeByPath(node.Children(), path)
		if child != nil {
			return child
		}
	}
	return nil
}

func buildNodeTree(t Treeish, paths []string) (Nodes, error) {
	flatNodes := make(Nodes, len(paths))
	for i, p := range paths {
		st, _ := t.State(p)
		if st&NodeCollapsible == NodeCollapsible {
			st |= NodeCollapsed
		}
		flatNodes[i] = &pathNode{
			Path:  p,
			state: st,
		}
	}
	nodes := make(Nodes, 0)
	for _, n := range flatNodes {
		ppath, _ := path.Split(n.String())
		if parent := findNodeByPath(flatNodes, ppath); parent != nil {
			if p, ok := parent.(*pathNode); ok {
				p.Nodes = append(p.Nodes, n)
			}
		} else {
			//fmt.Fprintf(os.Stderr, "Parent not found: %s vs %s\n", path.Clean(n.Path), path.Clean(ppath))
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

func walk(m *Model) error {
	paths, err := m.t.Walk(m.view.h - 5)
	if err != nil {
		m.debug("Err: %s\n", err)
	}
	m.tree, err = buildNodeTree(m.t, paths)
	if err != nil {
		m.debug("Err: %s\n", err)
	}
	return nil
}

// Update is the Tea update function which binds keystrokes to pagination.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var err error
	needsWalk := false
	switch msg := msg.(type) {
	//case TreeMsg:
	//	m.tree = append(m.tree, DebugNode{string(msg)})
	case tea.KeyMsg:
		// TODO(marius): we can create a data type that can be passed to the model and would function as key mapping.
		//   So the dev can load the mapping from someplace.
		//   There can be one where we add all the Readline bindings for example.
		switch msg.String() {
		case "`":
			m.showDebug = !m.showDebug
		case "enter":
			err = m.ToggleExpand()
		case "up", "k":
			err = m.Prev(1)
		case "down", "j":
			err = m.Next(1)
			needsWalk = true
		case "pgup":
			err = m.Prev(m.view.h)
			needsWalk = true
		case "pgdown":
			err = m.Next(m.view.h)
			needsWalk = true
		case "q", "esc", "ctrl+q", "ctrl+c":
			return m, tea.Quit
		default:
			fmt.Fprintf(os.Stderr, "unknown key %s\n", msg.String())
		}
	case tea.WindowSizeMsg:
		m.view.h = msg.Height
		m.view.w = msg.Width
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %s\n", err)
	}
	if needsWalk {
		walk(m)
	}
	return m, nil
}

// View renders the pagination to a string.
func (m Model) View() string {
	return m.render()
}

func (m Model) renderNode(t Node, i int) string {
	style := defaultStyle
	annotation := ""
	padding := ""

	//level := len(strings.Split(t.Path, "/")) - 1
	_, name := path.Split(t.String())

	if len(t.Children()) > 0 && t.State()&NodeCollapsed == NodeCollapsed {
		annotation = SquaredMinus
	}
	if len(t.Children()) > 0 && t.State()&NodeCollapsible == NodeCollapsible {
		annotation = SquaredPlus
	}

	if i == m.view.pos {
		style = highlightStyle
	}

	/*
		for j := 1; j < level; j++ {
			padding += BoxDrawingsHorizontal
		}
	*/

	return style.Width(m.view.w).Render(fmt.Sprintf("%s %2s %s", padding, annotation, name))
}

func (m Model) render() string {
	//fmt.Fprintf(os.Stderr, "WxH %dx%d - t:b %d:%d nc: %d\n", m.view.w, m.view.h, top, bot, len(m.tree))
	cursor := m.Nodes()
	lines := make([]string, 0)

	lastLine := 0
	for _, n := range cursor {
		lines = append(lines, m.renderNode(n, lastLine))
		lastLine++
		if len(n.Children()) > 0 {
			for _, c := range n.Children() {
				lines = append(lines, m.renderNode(c, lastLine))
				lastLine++
			}
		}
	}
	return strings.Join(lines, "\n")
}

func clamp(v, low, high int) int {
	return min(high, max(low, v))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
