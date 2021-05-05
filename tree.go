package tree

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// NodeState is used for passing information from a Treeish element to the view itself
type NodeState int

const (
	// NodeCollapsed hints that the current node is collapsed
	NodeCollapsed = 1 << iota
	// NodeCollapsible hints that the current node can be collapsed
	NodeCollapsible
	NodeError
	NodeDebug

	NodeNone = 0
)

type Treeish interface {
	// Advance moves the Treeish to a new received path,
	// this can return a new Treeish instance at the new path, or perform some other function
	// for the cases where the path doesn't correspond to a Treeish object.
	// Specifically in the case of the filepath Treeish:
	// If a passed path parameter corresponds to a folder, it will return a new Treeish object at the new path
	// If the passed path parameter corresponds to a file, it returns a nil Treeish but can execute something else.
	Advance(string) (Treeish, error)
	// State returns the NodeState for the received path parameter
	// This is used when rendering the path in the tree view
	State(string) (NodeState, error)

	// Walk loads the elements of current Treeish and returns them as a flat list
	Walk(int) ([]string, error)
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
func (m *Model) err(err error) {
	if m.debugNodes == nil {
		m.debugNodes = make([]Node, 0)
	}
	m.debugNodes = append(m.debugNodes, debug(err))
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
	s := strings.Builder{}
	nodeS := fmt.Sprintf("Path: %s [%d]\n", n.Path, len(n.Children()))
	s.WriteString(nodeS)
	if len(n.Children()) > 0 {
		s.WriteString(fmt.Sprintf("%#v\n", n.Children()))
	}
	return s.String()
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
	lines     []string
}

func (m *Model) bottom() int {
	bot := min(m.view.top+m.view.h, m.view.h) - 1
	if m.Debug {
		return bot - len(m.debugNodes)
	}
	return bot
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

func (n Nodes) at(i int) Node {
	for j, p := range n {
		if j == i {
			return p
		}
		if p.Children() != nil {
			if nn := p.Children().at(i - j - 1); nn != nil {
				return nn
			}
		}
	}
	return nil
}

func (n Nodes) GoString() string {
	s := strings.Builder{}
	for i, nn := range n {
		s.WriteString(fmt.Sprintf(" %d => %#v\n", i, nn))
	}
	return s.String()
}

// Model is the Bubble Tea model for this user interface.
type Model struct {
	tree       Nodes
	debugNodes Nodes
	Debug      bool

	t    Treeish
	view viewport
}

func New(t Treeish) *Model {
	m := new(Model)
	m.t = t
	return m
}

func (m *Model) Children() Nodes {
	return m.tree
}

// ToggleExpand toggles the expanded state of the node pointed at by m.view.pos
func (m *Model) ToggleExpand() error {
	return nil
}

// Parent moves the whole Treeish to the parent node
func (m *Model) Parent() error {
	n := m.tree.at(0)
	if n == nil {
		return fmt.Errorf("invalid node at pos %d", m.view.pos)
	}
	parent := path.Dir(n.String())
	t, err := m.t.Advance(parent)
	if err != nil {
		return err
	}
	if m.t != nil {
		m.debug("going to parent: %s", parent)
		m.t = t
	}
	return nil
}

// Advance moves the whole Treeish to the node m.view.pos points at
func (m *Model) Advance() error {
	n := m.tree.at(m.view.pos)
	if n == nil {
		return fmt.Errorf("invalid node at pos %d", m.view.pos)
	}
	// TODO(marius): this behaviour needs to be moved to the Treeish interface, as all implementations
	//   will need to know that a node is being collapsed or expanded.
	if pn, ok := n.(*pathNode); ok {
		if pn.state&NodeCollapsed == NodeCollapsed {
			t, err := m.t.Advance(n.String())
			if err != nil {
				return err
			}
			if m.t != nil {
				m.debug("advancing to: %s", n.String())
				m.t = t
				m.view.pos = 0
			}
		}
		pn.state ^= NodeCollapsed
	}
	return nil
}

// Top moves the current position to the first element
func (m *Model) Top() error {
	m.view.pos = 0
	m.view.top = 0
	m.debug("Top: top %d, pos: %d", m.view.top, m.view.pos)
	return nil
}

// Bottom moves the current position to the last element
func (m *Model) Bottom() error {
	m.view.pos = visibleLines(m.tree)
	m.view.top = min(visibleLines(m.tree)-m.view.top, m.view.top)
	m.debug("Bottom: top %d, pos: %d", m.view.top, m.view.pos)
	return nil
}

func visibleLines(n Nodes) int {
	count := len(n)
	for _, nn := range n {
		count += len(nn.Children()) - 1
	}
	return count
}

// Prev moves the current position to the previous 'i'th element in the tree.
// If it's above the viewport we need to recompute the top
func (m *Model) Prev(i int) error {
	m.view.pos = clamp(m.view.pos-i, 0, visibleLines(m.tree))
	if m.view.pos < m.bottom() {
		m.view.top = clamp(m.view.top-i, 0, max(m.view.h, visibleLines(m.tree)-m.view.h)-2)
	}
	m.debug("Prev: top %d, pos: %d bot: %d", m.view.top, m.view.pos, m.bottom())
	return nil
}

// Next moves the current position to the next 'i'th element in the tree
// If it's below the viewport we need to recompute the top
func (m *Model) Next(i int) error {
	m.view.pos = clamp(m.view.pos+i, 0, visibleLines(m.tree))
	if m.view.pos > m.bottom() {
		m.view.top = clamp(m.view.top+i, 0, max(m.view.h, visibleLines(m.tree)-m.view.h)-2)
	}
	m.debug("Next: top %d, pos: %d bot: %d", m.view.top, m.view.pos, m.bottom())
	return nil
}

type Msg string

func (m *Model) init() tea.Msg {
	walk(m)
	return Msg("inited")
}

func (m *Model) Init() tea.Cmd {
	return m.init
}

func findNodeByPath(nodes Nodes, path string) Node {
	for _, node := range nodes {
		if filepath.Clean(node.String()) == filepath.Clean(path) {
			return node
		}
		if child := findNodeByPath(node.Children(), path); child != nil {
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
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

func walk(m *Model) error {
	paths, err := m.t.Walk(m.view.h)
	if err != nil {
		m.err(err)
	}
	m.tree, err = buildNodeTree(m.t, paths)
	if err != nil {
		m.err(err)
	}
	return nil
}

// Update is the Tea update function which binds keystrokes to pagination.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.debugNodes = m.debugNodes[:0]
	var err error
	needsWalk := false
	switch msg := msg.(type) {
	case Msg:
		m.debug(string(msg))
	case tea.KeyMsg:
		// TODO(marius): we can create a data type that can be passed to the model and would function as key mapping.
		//   So the dev can load the mapping from someplace.
		//   There can be one where we add all the Readline bindings for example.
		switch msg.String() {
		case "`":
			m.Debug = !m.Debug
		case "enter":
			err = m.Advance()
			needsWalk = true
		case "backspace":
			err = m.Parent()
			needsWalk = true
		case "home":
			err = m.Top()
			needsWalk = true
		case "end":
			err = m.Bottom()
			needsWalk = true
		case "up", "k":
			err = m.Prev(1)
			needsWalk = true
		case "upup", "kk":
			err = m.Prev(2)
			needsWalk = true
		case "upupup", "kkk":
			err = m.Prev(3)
			needsWalk = true
		case "down", "j":
			err = m.Next(1)
			needsWalk = true
		case "downdown", "jj":
			err = m.Next(2)
			needsWalk = true
		case "downdowndown", "jjj":
			err = m.Next(3)
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
			m.err(fmt.Errorf("unknown key %q", msg))
		}
	case tea.WindowSizeMsg:
		m.view.h = msg.Height
		m.view.w = msg.Width
		m.view.lines = make([]string, m.view.h)
	}
	if err != nil {
		m.err(err)
	}
	if needsWalk {
		walk(m)
		for i := range m.view.lines {
			m.view.lines[i] = ""
		}
	}
	return m, nil
}

// View renders the pagination to a string.
func (m Model) View() string {
	return m.render()
}

const (
	NodeFirstChild = 1 << iota
	NodeLastChild
)

func (m Model) renderDebugNode(t Node) string {
	style := DefaultStyle
	annotation := ""

	if t.State()&NodeDebug == NodeDebug {
		style = DebugStyle
		annotation = ">"
	}
	if t.State()&NodeError == NodeError {
		style = ErrStyle
		annotation = "!"
	}

	return style.Width(m.view.w).Render(fmt.Sprintf("%2s %s", annotation, t.String()))
}

func (m Model) renderNode(t Node, cur int, nodeHints, depth int) string {
	style := DefaultStyle
	annotation := ""
	padding := ""

	if t.State()&NodeCollapsible == NodeCollapsible {
		annotation = SquaredMinus
		if t.State()&NodeCollapsed == NodeCollapsed {
			annotation = SquaredPlus
		}
	}

	for i := 0; i < depth; i++ {
		padding += "   " // 3 is the length of a tree opener
	}
	if nodeHints&NodeFirstChild == NodeFirstChild {
		padding += BoxDrawingsUpAndRight
	} else if nodeHints&NodeLastChild == NodeLastChild {
		padding += BoxDrawingsUpAndRight
	} else {
		padding += BoxDrawingsVerticalAndRight
	}
	padding += BoxDrawingsHorizontal

	if t.State()&NodeDebug == NodeDebug {
		style = DebugStyle
	}
	if t.State()&NodeError == NodeError {
		style = ErrStyle
	}

	if cur == m.view.pos+m.view.top {
		style = style.Reverse(true)
	}

	_, name := path.Split(t.String())
	return style.Width(m.view.w).Render(fmt.Sprintf("%s%2s %s", padding, annotation, name))
}

func (m Model) render() string {
	if m.view.h == 0 {
		return ""
	}
	cursor := m.Children()
	if cursor.Len() == 0 {
		return ""
	}
	//m.debug("WxH %dx%d - nc: %d", m.view.w, m.view.h, cursor.Len())

	maxLines := m.view.h
	if m.Debug {
		maxLines -= m.debugNodes.Len()
	}
	m.debug("displaying lines: t:%d b:%d tot:%d h:%d", m.view.top, maxLines, visibleLines(cursor), m.view.h)
	hints := NodeFirstChild
	for i := range m.view.lines {
		lineIndx := i + m.view.top
		if lineIndx >= cursor.Len() {
			break
		}
		n := cursor.at(lineIndx)
		if n == nil {
			continue
		}
		m.view.lines[i] = m.renderNode(n, lineIndx, hints, 0)
		hints = 0

		if childLen := len(n.Children()); childLen > 0 {
			for k, c := range n.Children() {
				lineIndx = i + k + 1
				if lineIndx >= maxLines {
					break
				}
				if k == childLen {
					hints |= NodeLastChild
				}
				m.view.lines[lineIndx] = m.renderNode(c, lineIndx, hints, 1)
				hints = 0
			}
		}
		if i > m.view.h-visibleLines(m.debugNodes) {
			break
		}
	}
	debStart := len(m.view.lines) - len(m.debugNodes)
	if m.Debug {
		for i, n := range m.debugNodes {
			lineIndx := debStart + i
			m.view.lines[lineIndx] = m.renderDebugNode(n)
		}
	}
	return strings.Join(m.view.lines, "\n")
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
